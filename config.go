// Package confy is a simple JSON configuration loader.
//
// Confy tries to not get in your way and accepts any kind of
// structure to load configuration into. Confy uses encoding/json
// to encode/decode JSON to/from your structure.
package confy

import (
	"encoding/json"
	"fmt"
)

// TODO: More documentation, examples and tests.

// Defaulter is an interface that can be implemented by a Configer
// to allow for default values to be set.
type Defaulter interface {
	// Default is called on a Defaulter to fill the Defaulter
	// with the default values. This is done before Unmarshalling
	// into the Defaulter.
	Default()
}

// Configer is a named interface{} for clarity
type Configer interface{}

// ConfigError is the error returned by local errors.
type ConfigError struct {
	Config      *Config
	Description string
}

func (err *ConfigError) Error() string {
	return err.Description
}

// Config is a configuration manager
type Config struct {
	Configer
	Subs map[string]*Config
	Raws map[string]*json.RawMessage
}

// NewConfig returns a new configuration manager, using c as its
// root Configer.
func NewConfig(c Configer) *Config {
	return &Config{
		Configer: c,
		Subs:     map[string]*Config{},
		Raws:     map[string]*json.RawMessage{},
	}
}

// AddSub adds a Configer to the current Config under the name given.
// Returns a *ConfigError if the name is already in use.
func (c *Config) AddSub(name string, sub Configer) error {
	return c.AddSubConf(name, NewConfig(sub))
}

// AddSubConf adds a Config to the current Config under the name given.
// Returns a *ConfigError if the name is already in use.
func (c *Config) AddSubConf(name string, sub *Config) error {
	existing, ok := c.Subs[name]
	if ok {
		return &ConfigError{
			Config: c,
			Description: fmt.Sprintf(
				"Sub configuration already exists with name '%s' (object: %v)",
				name, existing,
			),
		}
	}

	c.Subs[name] = sub
	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, c.Configer); err != nil {
		return err
	}

	if err := json.Unmarshal(b, &c.Raws); err != nil {
		return err
	}

	for name, raw := range c.Raws {
		sub, ok := c.Subs[name]
		if !ok {
			continue
		}

		if defaulter, ok := sub.Configer.(Defaulter); ok {
			defaulter.Default()
		}

		if err := json.Unmarshal(*raw, sub); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) MarshalJSON() ([]byte, error) {
	fullRaw, err := json.Marshal(c.Configer)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*json.RawMessage, 16)

	if err := json.Unmarshal(fullRaw, &m); err != nil {
		return nil, err
	}

	// Naive shallow copying of new values, over the old values
	for key, value := range m {
		c.Raws[key] = value
	}

	for name, sub := range c.Subs {
		raw, err := json.Marshal(sub)
		if err != nil {
			return nil, err
		}
		rawm := json.RawMessage(raw)
		c.Raws[name] = &rawm
	}

	return json.Marshal(c.Raws)
}
