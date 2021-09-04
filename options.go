package main

import "go.uber.org/zap"

// Options is a StickerBot options structure.
type Options struct {
	Logger *zap.Logger
}

func (cfg *Options) setDefaults() {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
}