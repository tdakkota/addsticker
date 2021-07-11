package main

import (
	"context"
	"io"
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type Stickers struct {
	user     *tg.InputPeerUser
	userMux  sync.Mutex
	signalCh chan struct{}
	sender   *message.Sender
}

func StickerBot(client *telegram.Client, dispatcher tg.UpdateDispatcher) *Stickers {
	s := &Stickers{
		signalCh: make(chan struct{}),
		sender:   message.NewSender(client.API()),
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

		if s.user == nil || s.user.UserID != peerID.UserID {
			return nil
		}

		select {
		case s.signalCh <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})

	return s
}

func (s *Stickers) getStickers(ctx context.Context) (*tg.InputPeerUser, error) {
	s.userMux.Lock()
	defer s.userMux.Unlock()

	if s.user != nil {
		return s.user, nil
	}

	u, err := s.sender.Resolve("@Stickers").AsInputUser(ctx)
	if err != nil {
		return nil, xerrors.Errorf("resolve: %w", err)
	}
	s.user = &tg.InputPeerUser{
		UserID:     u.UserID,
		AccessHash: u.AccessHash,
	}

	return s.user, nil
}

func (s *Stickers) Add(ctx context.Context, stickerPack, emoji string, sticker io.Reader) error {
	p, err := s.getStickers(ctx)
	if err != nil {
		return xerrors.Errorf("get Stickers peer: %w", err)
	}

	file, err := s.sender.To(p).Upload(message.FromReader("file.png", sticker)).AsInputFile(ctx)
	if err != nil {
		return xerrors.Errorf("upload: %w", err)
	}

	if err := s.send(ctx, "/cancel", "/addsticker", stickerPack); err != nil {
		return xerrors.Errorf("prepare: %w", err)
	}

	if err := s.sendImage(ctx, file); err != nil {
		return xerrors.Errorf("send sticker: %w", err)
	}

	if err := s.send(ctx, emoji, "/done"); err != nil {
		return xerrors.Errorf("send emoji %q: %w", emoji, err)
	}

	return nil
}

func (s *Stickers) await(ctx context.Context) error {
	select {
	case <-s.signalCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Stickers) send(ctx context.Context, texts ...string) error {
	p, err := s.getStickers(ctx)
	if err != nil {
		return xerrors.Errorf("get Stickers peer: %w", err)
	}

	for _, text := range texts {
		_, err := s.sender.To(p).Text(ctx, text)
		if err != nil {
			return xerrors.Errorf("send %q: %w", text, err)
		}

		if err := s.await(ctx); err != nil {
			return xerrors.Errorf("await %q: %w", text, err)
		}
	}
	return nil
}

func (s *Stickers) sendImage(ctx context.Context, file tg.InputFileClass) error {
	p, err := s.getStickers(ctx)
	if err != nil {
		return xerrors.Errorf("get Stickers peer: %w", err)
	}

	_, err = s.sender.To(p).Media(ctx, message.File(file).Filename("troll.png").MIME("image/png"))
	if err != nil {
		return xerrors.Errorf("send image: %w", err)
	}

	if err := s.await(ctx); err != nil {
		return xerrors.Errorf("await image: %w", err)
	}

	return nil
}
