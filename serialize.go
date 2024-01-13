package main

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

func serialize(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return errors.Wrap(err, "failed to marshal json")
	}
	return nil
}
