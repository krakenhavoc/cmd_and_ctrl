package lobby

// creator.go: game ownership (ADR 0051 decision 2, games.created_by),
// distinct from host.go's per-seat table host (ADR 0075 §2.1). A
// game's creator is whoever called POST /games while signed in; they
// need not ever sit down at the table, and a seated host need not be
// the creator. #1098 is the first thing to read it: the creator may
// rotate their own game's invites (POST /games/{id}/invites/rotate),
// same as the admin, and the Discord bot's /c2-end host check
// (GET /games/{id}/creator) answers a boolean against it without
// naming the creator to anyone but an admin caller who already
// guessed right.

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// ErrNotInviteManager is returned by rotateInvite's authorization
// check when the caller is neither the admin nor the game's creator.
var ErrNotInviteManager = errors.New("lobby: only the game's creator or the admin may rotate its invites")

// CanRotateInvites reports whether p may rotate meta's invites: the
// server admin, or the signed-in user who created this game
// (meta.CreatedBy). A guest, an unrelated signed-in user, or the
// creator of a DIFFERENT game is refused. A game with no creator
// (meta.CreatedBy == uuid.Nil — admin-created, or restored from a
// pre-ADR-0051 file import) can never match: p.UserID is checked
// non-zero first, so an admin principal (whose UserID is always Nil)
// never spuriously matches a Nil creator via the admin branch above
// it, and a Nil-UserID guest never matches one either.
func CanRotateInvites(p auth.Principal, meta GameMeta) bool {
	if p.Role == auth.RoleAdmin {
		return true
	}
	return p.UserID != uuid.Nil && meta.CreatedBy != uuid.Nil && p.UserID == meta.CreatedBy
}
