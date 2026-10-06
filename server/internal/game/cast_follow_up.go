package game

import (
	"sync"

	"github.com/google/uuid"
)

// cast_follow_up.go — #2173: something that happens BECAUSE a spell was
// cast through a cast permission.
//
// Conduit of Worlds lets you cast a card from your graveyard "if you
// haven't cast a spell this turn", and "if you do, you can't cast
// additional spells this turn". The second sentence is about the cast
// having happened, so it cannot be written beside the grant (which
// would bind a player who declined) or ahead of it (which would stop
// the very cast). It runs when the permission is used.
//
// The follow-up is DATA on the permission: a key into the registry
// below, the shape RegisterDrawThen (#2388) uses for "draw, then X".
// CastPermission is persisted and mirrored by the snapshot, so it holds
// no closure; the body is registered once at init by the card package
// and looked up by key when the cast is made.

// CastFollowUpKey names a registered CastFollowUpFunc. The empty key is
// "nothing follows".
type CastFollowUpKey string

// CastFollowUp is what a follow-up body is told: who cast, which
// permission source granted the cast, and the spell that is now on the
// stack. Plain data.
type CastFollowUp struct {
	Player uuid.UUID
	Source uuid.UUID
	Spell  uuid.UUID
}

// CastFollowUpFunc is the body of a follow-up. It runs with g.mu held,
// after the spell is on the stack and EventCast has been emitted. It
// must capture nothing that lives in the game.
type CastFollowUpFunc func(g *Game, f CastFollowUp) error

var (
	castFollowUpMu     sync.RWMutex
	castFollowUpBodies = map[string]CastFollowUpFunc{}
)

// RegisterCastFollowUp registers a follow-up body under key and returns
// the key to put on CastPermission.FollowUp. Idempotent for one key.
func RegisterCastFollowUp(key string, body CastFollowUpFunc) CastFollowUpKey {
	if key == "" || body == nil {
		panic("game.RegisterCastFollowUp: a key and a body are required")
	}
	castFollowUpMu.Lock()
	castFollowUpBodies[key] = body
	castFollowUpMu.Unlock()
	return CastFollowUpKey(key)
}

// runCastFollowUpLocked runs a registered follow-up. An unknown key (a
// snapshot from a build that registered more) is a no-op, never a
// panic.
//
// Caller must hold g.mu (write).
func (g *Game) runCastFollowUpLocked(key CastFollowUpKey, f CastFollowUp) error {
	if key == "" {
		return nil
	}
	castFollowUpMu.RLock()
	body := castFollowUpBodies[string(key)]
	castFollowUpMu.RUnlock()
	if body == nil {
		return nil
	}
	return body(g, f)
}
