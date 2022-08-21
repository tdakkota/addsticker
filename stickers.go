package main

import (
	"context"
	"io"
	"sync"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// Stickers is a simple helper for interaction with @Stickers bot.
type Stickers struct {
	// Resolved @Stickers peer.
	user    *tg.InputPeerUser
	userMux sync.Mutex
	// Message channel.
	signalCh chan *tg.Message

	sender *message.Sender
	log    *zap.Logger
}

// StickerBot creates new Stickers and sets hooks to given dispatcher.
func StickerBot(client *telegram.Client, dispatcher tg.UpdateDispatcher, opts Options) *Stickers {
	opts.setDefaults()
	s := &Stickers{
		signalCh: make(chan *tg.Message),
		sender:   message.NewSender(client.API()),
		log:      opts.Logger,
	}
	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}

		peerID, ok := msg.PeerID.(*tg.PeerUser)
		if !ok {
			return nil
		}

		stickers, err := s.getStickers(ctx)
		if err != nil {
			return errors.Wrap(err, "get Stickers")
		}

		if stickers.UserID != peerID.UserID {
			return nil
		}

		select {
		case s.signalCh <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})

	return s
}

// getStickers resolves @Stickers once.
func (s *Stickers) getStickers(ctx context.Context) (*tg.InputPeerUser, error) {
	s.userMux.Lock()
	defer s.userMux.Unlock()

	if s.user != nil {
		return s.user, nil
	}

	u, err := s.sender.Resolve("@Stickers").AsInputUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "resolve")
	}
	s.user = &tg.InputPeerUser{
		UserID:     u.UserID,
		AccessHash: u.AccessHash,
	}

	return s.user, nil
}

// Add adds sticker to stickerPack using given emoji list and sticker.
func (s *Stickers) Add(ctx context.Context, stickerPack, emoji string, sticker io.Reader) error {
	p, err := s.getStickers(ctx)
	if err != nil {
		return errors.Wrap(err, "get Stickers peer")
	}

	file, err := s.sender.To(p).Upload(message.FromReader("file.png", sticker)).AsInputFile(ctx)
	if err != nil {
		return errors.Wrap(err, "upload")
	}

	if err := s.send(ctx, "/cancel", "/addsticker", stickerPack); err != nil {
		return errors.Wrap(err, "prepare")
	}

	if err := s.sendImage(ctx, file); err != nil {
		return errors.Wrap(err, "send sticker")
	}

	if err := s.send(ctx, emoji, "/done"); err != nil {
		return errors.Wrapf(err, "send emoji %q", emoji)
	}

	return nil
}

// await waits for answer from bot.
//
//nolint:unparam
func (s *Stickers) await(ctx context.Context) (*tg.Message, error) {
	select {
	case msg := <-s.signalCh:
		s.log.Info("Received message", zap.String("message", msg.Message))
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// send sends every text as new message and awaits result.
func (s *Stickers) send(ctx context.Context, texts ...string) error {
	p, err := s.getStickers(ctx)
	if err != nil {
		return errors.Wrap(err, "get Stickers peer")
	}

	for _, text := range texts {
		_, err := s.sender.To(p).Text(ctx, text)
		if err != nil {
			return errors.Wrapf(err, "send %q", text)
		}
		s.log.Info("Sent message", zap.String("message", text))

		if _, err := s.await(ctx); err != nil {
			return errors.Wrapf(err, "await %q", text)
		}
	}
	return nil
}

// sendImage sends image and awaits result.
func (s *Stickers) sendImage(ctx context.Context, file tg.InputFileClass) error {
	p, err := s.getStickers(ctx)
	if err != nil {
		return errors.Wrap(err, "get Stickers peer")
	}

	_, err = s.sender.To(p).Media(ctx, message.File(file).Filename("sticker.png").MIME("image/png"))
	if err != nil {
		return errors.Wrap(err, "send image")
	}

	if _, err := s.await(ctx); err != nil {
		return errors.Wrap(err, "await image")
	}

	return nil
}
