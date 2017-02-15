package helper

import (
	"os"

	"github.com/pkg/errors"
)

func mkdir(newDir string) error {
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		err = os.Mkdir(newDir, 0750)
		if err != nil {
			return errors.Wrap(err, "Failed to mkdir")
		}
	}
	return nil
}
