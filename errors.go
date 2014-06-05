package congo

import (
	"encoding/json"
	"fmt"
)

// ConfigError is the error returned when an error occurs in
// non-json congo logic.
type ConfigError struct {
	Config      *Config
	Description string
}

func (err *ConfigError) Error() string {
	return err.Description
}

// PrettifyError tries to make an error returned by congo more
// human readable.
//
// FIXME: This currently doesn't do much. A placeholder.
func PrettifyError(err error) string {
	switch u := err.(type) {
	case *json.SyntaxError:
		return fmt.Sprintf(`syntax error at offset %d: %s`,
			u.Offset, u)
	case *json.UnmarshalTypeError:
		return "congo: " + err.Error()
	case *json.InvalidUnmarshalError:
		return "congo: " + err.Error()
	default:
		return "congo: " + err.Error()
	}
}
