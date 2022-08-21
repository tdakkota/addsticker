package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func run(ctx context.Context) error {
	if err := tryLoadEnv(); err != nil {
		return errors.Wrap(err, "load env")
	}

	var (
		alt, imagePath, pack string
		log                  bool
	)
	flag.StringVar(&alt, "emoji", "", "emoji to add")
	flag.StringVar(&imagePath, "image", "", "image to add")
	flag.StringVar(&pack, "pack", "wtfakkota2", "pack to add")
	flag.BoolVar(&log, "log", false, "enable logging")
	flag.Parse()

	if alt == "" || imagePath == "" {
		return errors.New("emoji or image arg is empty")
	}

	logger := zap.NewNop()
	if log {
		l, err := zap.NewDevelopment()
		if err != nil {
			return errors.Wrap(err, "create logger")
		}
		logger = l
	}
	defer func() {
		_ = logger.Sync()
	}()

	dispatcher := tg.NewUpdateDispatcher()
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		UpdateHandler: dispatcher,
		Middlewares:   []telegram.Middleware{backoffRetry},
		Device: telegram.DeviceConfig{
			SystemLangCode: "en-US",
			LangCode:       "en",
		},
	})
	if err != nil {
		return errors.Wrap(err, "create client")
	}

	return client.Run(ctx, func(ctx context.Context) (rErr error) {
		f, err := os.Open(filepath.Clean(imagePath))
		if err != nil {
			return errors.Wrap(err, "open sticker")
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))

		return StickerBot(client, dispatcher, Options{Logger: logger.Named("stickerbot")}).
			Add(ctx, pack, alt, f)
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())

		if e, ok := errors.Into[*ExitError](err); ok {
			os.Exit(e.Code)
		}

		os.Exit(2)
	}
}
