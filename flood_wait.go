package main

import (
	"context"

	"github.com/cenkalti/backoff/v4"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

//nolint:gochecknoglobals
var backoffRetry telegram.MiddlewareFunc = func(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		return backoff.Retry(func() error {
			if err := next.Invoke(ctx, input, output); err != nil {
				if ok, err := tgerr.FloodWait(ctx, err); ok {
					return err
				}

				return backoff.Permanent(err)
			}

			return nil
		}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	}
}
