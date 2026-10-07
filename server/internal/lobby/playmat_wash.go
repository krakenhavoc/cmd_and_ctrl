package lobby

// playmat_wash.go: the owner-set wash beside each seat's playmat (ADR
// 0128 amendment). The source is the same *playmat.Service; it is asked
// through its own small interface so a PlaymatSource that has no wash
// (a test fake) still works, and every seat then shows the default.

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// PlaymatWashSource answers an account's wash. *playmat.Service
// implements it.
type PlaymatWashSource interface {
	Wash(user uuid.UUID) int
}

// bindPlaymatWash puts userID's wash on seat playerID of room, beside
// the URL bindPlaymat put there. Same locking rules as bindPlaymat.
func (l *Lobby) bindPlaymatWash(room *ws.Room, playerID, userID uuid.UUID) {
	if l.playmats == nil || userID == uuid.Nil || room == nil {
		return
	}
	src, ok := l.playmats.(PlaymatWashSource)
	if !ok {
		return
	}
	room.SetPlaymatWash(playerID, src.Wash(userID))
}
