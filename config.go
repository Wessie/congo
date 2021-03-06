package congo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// BUG(Wessie): Anything not JSON-object will error as root of a tree.

// Defaulter is an interface that can be implemented by a Configer
// to allow for default values to be set.
type Defaulter interface {
	// Default is called on a Defaulter to fill the value with
	// the default values. This is done before any Unmarshal
	// calls are made on the value.
	Default()
}

// Configer is a named interface{} for clarity
type Configer interface{}

// Config is a configuration manager
type Config struct {
	Loaded bool
	Configer
	Subs map[string]*Config
	Raws map[string]*json.RawMessage
}

// NewConfig returns a new configuration manager, using c as its
// root.
func NewConfig(c Configer) *Config {
	return &Config{
		Configer: c,
		Subs:     map[string]*Config{},
		Raws:     map[string]*json.RawMessage{},
	}
}

// DefaultAll calls Default on c.Configer and any Defaulters in c.Subs
func (c *Config) DefaultAll() {
	c.Loaded = true
	if d, ok := c.Configer.(Defaulter); ok {
		d.Default()
	}

	for _, sub := range c.Subs {
		sub.DefaultAll()
	}
}

// AddSub adds a Configer to the current Config under the name given.
// See AddSubConf for extra behaviour, this is a wrapper of it.
func (c *Config) AddSub(name string, sub Configer) error {
	return c.AddSubConf(name, NewConfig(sub))
}

func AddSub(name string, sub Configer) error {
	return Default.AddSub(name, sub)
}

// AddSubConf adds a Config to the current Config under the name given.
//
// If a configuration was already loaded before the call to AddSubConf the
// new Config will be filled with any data from the previous load.
//
// Returns a *ConfigError if the name is already in use or a JSON error if
// something went wrong unmarshalling.
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

	// Check if we've already parsed JSON, if so check for the new name and
	// unmarshal it into the new config we just registered.
	if c.Loaded {
		raw, ok := c.Raws[name]
		if !ok {
			// It isn't an error if we have nothing for the new config
			// except some defaults.
			if d, ok := sub.Configer.(Defaulter); ok {
				d.Default()
			}
			return nil
		}

		// Is an error if unmarshalling borks, so return that
		return json.Unmarshal(*raw, sub)
	}
	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	c.Loaded = true

	// TODO: This is wasteful due to the recursive calls
	c.DefaultAll()

	if c.Configer == nil {
	} else if err := json.Unmarshal(b, c.Configer); err != nil {
		return err
	}

	if err := json.Unmarshal(b, &c.Raws); err != nil {
		return err
	}

	for name, sub := range c.Subs {
		raw, ok := c.Raws[name]
		if !ok {
			continue
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

	// TODO: Don't assume map or struct here
	// We assume our Configer is a map or struct here, this is sub-optimal but
	// not a breaking problem.
	m := make(map[string]*json.RawMessage, 16)
	if err := json.Unmarshal(fullRaw, &m); err != nil {
		return nil, err
	}

	// If our Configer is a Defaulter we remove any key/value pairs that are set to
	// their default value to avoid marshalling them to JSON.
	if defaulter, ok := newDefaulter(c.Configer); ok {
		defaulter.Default()

		raw, err := json.Marshal(defaulter)
		if err != nil {
			return nil, err
		}

		dm := make(map[string]*json.RawMessage, len(m))
		if err := json.Unmarshal(raw, &dm); err != nil {
			return nil, err
		}

		for key, defaultValue := range dm {
			value, ok := m[key]
			if !ok {
				// We don't need to remove things that don't exist
				continue
			}

			if bytes.Equal(*value, *defaultValue) {
				delete(m, key)
			}
		}
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

		// Check if a sub returned an empty object or array, and omit
		// them if so. It being empty means default values were omitted.
		if len(raw) == 2 && (raw[0] == '{' || raw[0] == '[') {
			continue
		}

		rawm := json.RawMessage(raw)
		c.Raws[name] = &rawm
	}

	return json.Marshal(c.Raws)
}

// newDefaulter returns a new Configer if it is a Defaulter.
// d will be non-nil only if ok is true. The Defaulter is
// the zero-value of whatever type Configer is.
func newDefaulter(c Configer) (d Defaulter, ok bool) {
	d, ok = c.(Defaulter)
	if !ok {
		return d, ok
	}

	v := reflect.ValueOf(c)

	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		return nil, false
	}

	n := reflect.New(v.Elem().Type())

	d, ok = n.Interface().(Defaulter)
	return d, ok
}
