package main

import (
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
	"github.com/joho/godotenv"
)

func tryLoadEnv() error {
	tryOne := func(p string) (bool, error) {
		_, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, errors.Wrap(err, "stat")
		}

		return true, godotenv.Load(p)
	}

	load, err := tryOne(".tdenv")
	if load || err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get home")
	}
	repl := filepath.Join(home, ".td", "tdrepl.env")

	_, err = tryOne(repl)
	return err
}
