package validation

import "fmt"

var allowed = map[string]bool{"received": true, "cold": true, "quarantine": true, "released": true, "recalled": true}

func Status(v string) error {
	if !allowed[v] {
		return fmt.Errorf("unsupported status %q", v)
	}
	return nil
}
