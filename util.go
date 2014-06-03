package confy

import (
	"encoding/json"
	"io"
)

var Default = NewConfig(nil)

// LoadReader calls Config.LoadReader on the Default config
func LoadReader(r io.Reader) error {
	return Default.LoadReader(r)
}

// LoadReader loads the configuration from r into c.
func (c *Config) LoadReader(r io.Reader) error {
	return json.NewDecoder(r).Decode(Default)
}

// SaveWriter calls Config.SaveWriter on the Default config
func SaveWriter(w io.Writer) error {
	return Default.SaveWriter(w)
}

// SaveWriter saves the configuration c into w. The result
// is human-readable indented before writing.
func (c *Config) SaveWriter(w io.Writer) error {
	b, err := json.Marshal(Default)
	if err != nil {
		return err
	}

	for n, nn := 0, 0; err == nil && n < len(b); n += nn {
		nn, err = w.Write(b)
	}

	return err
}
