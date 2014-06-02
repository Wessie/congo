// Package confy is a simple JSON configuration loader.
//
// Confy uses the standard library encoding/json package and
// supports the same things as it does.
//
// TODO: More documentation, examples and tests.
package confy

import "encoding/json"

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

// AddSub adds a Config to the current Config under the name given.
//
// If previously a Config was already added by the name given, AddSub
// will return said Config instead of sub. This can never return nil.
func (c *Config) AddSub(name string, sub *Config) *Config {
	existing, ok := c.Subs[name]
	if ok {
		return existing
	}

	c.Subs[name] = sub
	return sub
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
