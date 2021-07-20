package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type ExitError struct {
	Msg  string
	Code int
}

func (e *ExitError) Error() string {
	return e.Msg
}

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

func run(ctx context.Context) error {
	if err := tryLoadEnv(); err != nil {
		return xerrors.Errorf("load env: %w", err)
	}

	var (
		alt, imagePath, pack string
	)
	flag.StringVar(&alt, "emoji", "", "emoji to add")
	flag.StringVar(&imagePath, "image", "", "image to add")
	flag.StringVar(&pack, "pack", "wtfakkota2", "pack to add")
	flag.Parse()

	if alt == "" || imagePath == "" {
		return xerrors.New("emoji or image arg is empty")
	}

	logger := zap.NewNop()
	defer func() {
		_ = logger.Sync()
	}()

	dispatcher := tg.NewUpdateDispatcher()
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		UpdateHandler: dispatcher,
		Middlewares:   []telegram.Middleware{backoffRetry},
	})
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) (rErr error) {
		f, err := os.Open(filepath.Clean(imagePath))
		if err != nil {
			return xerrors.Errorf("open sticker: %w", err)
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))

		return StickerBot(client, dispatcher).Add(ctx, pack, alt, f)
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())

		var e *ExitError
		if xerrors.As(err, &e) {
			os.Exit(e.Code)
		}

		os.Exit(2)
	}
}
