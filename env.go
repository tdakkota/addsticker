package main

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"golang.org/x/xerrors"
)

func tryLoadEnv() error {
	tryOne := func(p string) (bool, error) {
		_, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, xerrors.Errorf("stat: %w", err)
		}

		return true, godotenv.Load(p)
	}

	load, err := tryOne(".tdenv")
	if load || err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return xerrors.Errorf("get home: %w", err)
	}
	repl := filepath.Join(home, ".td", "tdrepl.env")

	_, err = tryOne(repl)
	return err
}
