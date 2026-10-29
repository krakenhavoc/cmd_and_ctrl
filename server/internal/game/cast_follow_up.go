package game

import (
	"sort"
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

	// Object is the permission's FollowUpObject: an object the "if you
	// do" acts on besides the spell (Spellweaver Volute's enchanted
	// card, #3157). Zero when the permission names none.
	Object ObjectRef
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

// KnownCastFollowUp reports whether a follow-up body is registered under
// key.
func KnownCastFollowUp(key CastFollowUpKey) bool {
	castFollowUpMu.RLock()
	defer castFollowUpMu.RUnlock()
	_, ok := castFollowUpBodies[string(key)]
	return ok
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

// unknownCastFollowUps is checkEffectKeys' line for each follow-up key a
// stored cast permission names that this binary does not register
// (#4361 S13): FollowUp and IfNotCast. runCastFollowUpLocked runs an
// unknown key as nothing, so a restore that kept one would grant the cast
// without its "if you do" or "if you don't"; the file is refused instead.
func unknownCastFollowUps(p *CastPermission) []string {
	if p == nil {
		return nil
	}
	var out []string
	for _, k := range []CastFollowUpKey{p.FollowUp, p.IfNotCast} {
		if k != "" && !KnownCastFollowUp(k) {
			out = append(out, "cast follow-up "+string(k))
		}
	}
	return out
}

// registeredCastFollowUpKeys lists the ledger lines, "castfollowup <key>"
// (#4361 S13): a stored cast permission names its follow-ups by key, so
// they are on-disk identities.
func registeredCastFollowUpKeys() []string {
	castFollowUpMu.RLock()
	defer castFollowUpMu.RUnlock()
	out := make([]string, 0, len(castFollowUpBodies))
	for k := range castFollowUpBodies {
		out = append(out, "castfollowup "+k)
	}
	sort.Strings(out)
	return out
}
