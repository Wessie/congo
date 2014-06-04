package congo

import (
	"encoding/json"
	"io"
	"os"
)

// Indentation indicates what to indent the JSON with when
// writing it out.
var Indentation = "\t"

// Default is a global config for when you don't need more
// than just that.
var Default = NewConfig(nil)

// SetRoot sets the root of the Default configuration.
func SetRoot(conf Configer) {
	Default.SetRoot(conf)
}

// SetRoot sets the root Configer in the configuration
func (c *Config) SetRoot(conf Configer) {
	c.Configer = conf
}

// LoadReader calls Config.LoadReader on the Default config
func LoadReader(r io.Reader) error {
	return Default.LoadReader(r)
}

type countedReader struct {
	io.Reader
	n int
}

func (r *countedReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.n += n
	return n, err
}

// LoadReader loads the configuration from r into c.
//
// If the reader passed in contains no data the configuration
// will be set to it's default value with the aid of Defaulter's.
func (c *Config) LoadReader(r io.Reader) error {
	rc := &countedReader{Reader: r}
	err := json.NewDecoder(rc).Decode(c)

	// Special case an io.EOF if we have read 0 bytes. It makes
	// no sense to error in the case of an empty configuration file.
	if err == io.EOF && rc.n == 0 {
		c.DefaultAll()
		return nil
	}

	return err
}

// LoadFile loads a configuration from the file indicated by path.
// For more control of loading see LoadReader.
func (c *Config) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	return c.LoadReader(f)
}

// SaveWriter calls Config.SaveWriter on the Default config
func SaveWriter(w io.Writer) error {
	return Default.SaveWriter(w)
}

// SaveWriter saves the configuration c into w. The result
// is human-readable indented.
func (c *Config) SaveWriter(w io.Writer) error {
	b, err := json.MarshalIndent(c, "", Indentation)
	if err != nil {
		return err
	}

	for n, nn := 0, 0; err == nil && n < len(b); n += nn {
		nn, err = w.Write(b)
	}

	return err
}

// SaveFile saves a configuration into the file indicated by path.
//
// This calls SaveWriter underneath and the same indenting applies.
func (c *Config) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return c.SaveWriter(f)
}
