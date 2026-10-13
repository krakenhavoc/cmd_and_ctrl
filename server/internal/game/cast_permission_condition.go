package game

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// cast_permission_condition.go — #3192, ADR 0066's 2026-10-13 second
// amendment: a granted permission that is live only while a CONDITION
// holds.
//
// "During any turn you attacked with a token, you may play that card"
// (Neyali), "you may play it if you control a Kavu" (Possibility
// Technician), "during any turn you put a lore counter on this Saga"
// (Summon: Brynhildr). The permission outlives the turn it was granted
// in and is not a duration: it is dormant on the turns the condition is
// false and live on the turns it is true, and the card stays exiled the
// whole time. Before this, those cards were re-granted by a trigger that
// had to resolve on the right turn, which is weaker than printed.
//
// WHERE IT IS ASKED. CastPermissionActiveForEffect is THE liveness test
// every consumer reads (the cast path, the view's castable flags and
// exile_play stamp, the legal-move enumerator and so the bot, the
// library-top look and the hand branch), so the condition is evaluated
// there and nowhere else. It cannot hold for the engine and fail for the
// bot, which is the property #3105's cardOpensZoneLocked gave a card's
// OWN conditional zone; this is the same idea for a GRANTED permission,
// which is per-instance state and therefore cannot carry a closure.
//
// WHAT IT IS. A key into the registry below plus plain parameters
// (Subtype, N, an object), so it survives Clone, undo and a snapshot,
// and a snapshot naming a key this binary never registered is refused
// at restore rather than restored as an unconditional grant. An unknown
// key at run time fails CLOSED: the permission is dormant, never
// unconditional.
//
// WHAT IT IS NOT. A spell-type gate ("if it's a creature spell", "if
// it's an instant or sorcery spell") is not a game-state condition at
// all; it is a property of the card being cast, so it rides the
// permission's Filter, which CoversCard now applies to a ScopeCards
// permission as well. And "for as long as you control <source>" is
// already a source-bound Duration (ForAsLongAs + WhileYouControlSource),
// read by the same liveness test; nothing here duplicates it.

// PermissionConditionKey names a registered condition.
type PermissionConditionKey string

// The built-in conditions. Each is registered in this file's init, so a
// card never has to.
const (
	// CondAttackedWithToken: the holder attacked with a token this turn
	// (Neyali, Suns' Vanguard). Judged as the creature was DECLARED, so a
	// token that has since died still counts.
	CondAttackedWithToken PermissionConditionKey = "attacked_with_token"
	// CondAttackedWithCommander: the holder attacked with a commander
	// this turn (Neriv, Crackling Vanguard).
	CondAttackedWithCommander PermissionConditionKey = "attacked_with_commander"
	// CondAttackedWithSubtype: attacked with a creature that had
	// Subtype as it attacked (Robber of the Rich's Rogue). A changeling
	// has every creature type.
	CondAttackedWithSubtype PermissionConditionKey = "attacked_with_subtype"
	// CondAttackedWithAtLeast: attacked with N or more different
	// creatures this turn (Boros Strike-Captain's three).
	CondAttackedWithAtLeast PermissionConditionKey = "attacked_with_at_least"
	// CondAttackedWithObject: attacked with one specific object this
	// turn (Goblin Researcher's "this creature"). The object is its
	// ObjectID and the CR 400.7 epoch it had when the grant was made, so
	// a creature that left and came back is a different one.
	CondAttackedWithObject PermissionConditionKey = "attacked_with_object"
	// CondAttackedThisTurn: attacked with any creature this turn (Pick
	// Up the Pace's "as long as you attacked this turn").
	CondAttackedThisTurn PermissionConditionKey = "attacked_this_turn"
	// CondControlSubtype: the holder controls a permanent with Subtype
	// right now (Possibility Technician's Kavu, Gandalf's Wizard).
	CondControlSubtype PermissionConditionKey = "control_subtype"
	// CondLoreCounterOnObject: a lore counter was put on one specific
	// Saga this turn (Summon: Brynhildr's "this Saga").
	CondLoreCounterOnObject PermissionConditionKey = "lore_counter_on_object"
)

// PermissionCondition is a registered condition and its plain
// parameters. Which parameters a key reads is its registration's
// business; the rest are zero.
type PermissionCondition struct {
	Key         PermissionConditionKey `json:"key"`
	Subtype     string                 `json:"subtype,omitempty"`
	N           int                    `json:"n,omitempty"`
	ObjectID    uuid.UUID              `json:"objectID,omitempty"`
	ObjectEpoch int                    `json:"objectEpoch,omitempty"`
}

// PermissionConditionSpec is what a registration supplies: the
// predicate, asked of the permission's HOLDER, and the sentence the
// client prints for it. Both must be pure and capture nothing.
type PermissionConditionSpec struct {
	Holds func(g *Game, holder uuid.UUID, c *PermissionCondition) bool
	// Label is the clause as the card prints it, lower-case and without
	// a verb: "during a turn you attacked with a token".
	Label func(c *PermissionCondition) string
}

var (
	permissionConditionMu    sync.RWMutex
	permissionConditionSpecs = map[PermissionConditionKey]PermissionConditionSpec{}
)

// RegisterPermissionCondition registers a condition under key. Panics
// on an empty key or a missing predicate; a second registration under
// one key replaces the first (tests rely on that being harmless).
func RegisterPermissionCondition(key PermissionConditionKey, spec PermissionConditionSpec) PermissionConditionKey {
	if key == "" || spec.Holds == nil {
		panic("game.RegisterPermissionCondition: a key and a predicate are required")
	}
	permissionConditionMu.Lock()
	permissionConditionSpecs[key] = spec
	permissionConditionMu.Unlock()
	return key
}

// Known reports whether this binary registered the key. A restore point
// naming one that is not is refused (ErrUnknownEffectKey).
func (k PermissionConditionKey) Known() bool {
	permissionConditionMu.RLock()
	_, ok := permissionConditionSpecs[k]
	permissionConditionMu.RUnlock()
	return ok
}

func permissionConditionSpecFor(k PermissionConditionKey) (PermissionConditionSpec, bool) {
	permissionConditionMu.RLock()
	spec, ok := permissionConditionSpecs[k]
	permissionConditionMu.RUnlock()
	return spec, ok
}

// permissionConditionHoldsLocked is the one evaluation: true for a
// permission with no condition, the registered predicate's answer for
// one with, and false for a key nobody registered.
//
// Caller must hold g.mu (read or write).
func (g *Game) permissionConditionHoldsLocked(p *CastPermission, holder uuid.UUID) bool {
	if p == nil || p.Condition == nil {
		return true
	}
	spec, ok := permissionConditionSpecFor(p.Condition.Key)
	if !ok {
		return false
	}
	return spec.Holds(g, holder, p.Condition)
}

// PermissionConditionUnmetLocked reports whether the permission has a
// condition that is false right now, for the view: the card keeps its
// grant and the client greys the button. False for an unconditional
// permission.
//
// Caller must hold g.mu (read or write).
func (g *Game) PermissionConditionUnmetLocked(p *CastPermission) bool {
	return p != nil && p.Condition != nil && !g.permissionConditionHoldsLocked(p, p.Player)
}

// PermissionConditionLabel is the printed clause for the permission's
// condition, or "" when it has none.
func PermissionConditionLabel(p *CastPermission) string {
	if p == nil || p.Condition == nil {
		return ""
	}
	spec, ok := permissionConditionSpecFor(p.Condition.Key)
	if !ok || spec.Label == nil {
		return ""
	}
	return spec.Label(p.Condition)
}

func init() {
	RegisterPermissionCondition(CondAttackedWithToken, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, _ *PermissionCondition) bool {
			return g.TurnTally.Players[holder].TokensAttacked > 0
		},
		Label: func(*PermissionCondition) string { return "during a turn you attacked with a token" },
	})
	RegisterPermissionCondition(CondAttackedWithCommander, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, _ *PermissionCondition) bool {
			return g.TurnTally.Players[holder].CommandersAttacked > 0
		},
		Label: func(*PermissionCondition) string { return "during a turn you attacked with a commander" },
	})
	RegisterPermissionCondition(CondAttackedWithSubtype, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, c *PermissionCondition) bool {
			return c.Subtype != "" && g.AttackedWithSubtypeThisTurn(holder, c.Subtype) > 0
		},
		Label: func(c *PermissionCondition) string {
			return "during a turn you attacked with " + withArticle(c.Subtype)
		},
	})
	RegisterPermissionCondition(CondAttackedWithAtLeast, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, c *PermissionCondition) bool {
			return c.N > 0 && g.DistinctAttackersThisTurn(holder) >= c.N
		},
		Label: func(c *PermissionCondition) string {
			return fmt.Sprintf("during a turn you attacked with %d or more creatures", c.N)
		},
	})
	RegisterPermissionCondition(CondAttackedWithObject, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, c *PermissionCondition) bool {
			return c.ObjectID != uuid.Nil && g.AttackedWithObjectThisTurn(holder, c.ObjectID, c.ObjectEpoch)
		},
		Label: func(*PermissionCondition) string {
			return "during a turn you attacked with the creature that exiled it"
		},
	})
	RegisterPermissionCondition(CondAttackedThisTurn, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, _ *PermissionCondition) bool {
			return g.TurnTally.Players[holder].AttacksDeclared > 0
		},
		Label: func(*PermissionCondition) string { return "as long as you attacked this turn" },
	})
	RegisterPermissionCondition(CondControlSubtype, PermissionConditionSpec{
		Holds: func(g *Game, holder uuid.UUID, c *PermissionCondition) bool {
			return c.Subtype != "" && g.controlsSubtypeLocked(holder, c.Subtype)
		},
		Label: func(c *PermissionCondition) string { return "if you control " + withArticle(c.Subtype) },
	})
	RegisterPermissionCondition(CondLoreCounterOnObject, PermissionConditionSpec{
		Holds: func(g *Game, _ uuid.UUID, c *PermissionCondition) bool {
			return c.ObjectID != uuid.Nil && g.LoreCounterPutOnThisTurn(c.ObjectID, c.ObjectEpoch)
		},
		Label: func(*PermissionCondition) string { return "during a turn you put a lore counter on the Saga" },
	})
}

// withArticle prefixes "a" or "an" to a type name.
func withArticle(word string) string {
	if word == "" {
		return "a creature"
	}
	switch strings.ToLower(word[:1]) {
	case "a", "e", "i", "o", "u":
		return "an " + word
	}
	return "a " + word
}

// controlsSubtypeLocked reports whether player controls a battlefield
// permanent that has `subtype` right now (effective subtypes, so a
// changeling is every creature type).
//
// Caller must hold g.mu.
func (g *Game) controlsSubtypeLocked(player uuid.UUID, subtype string) bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && hasSubtype(c, subtype) {
			return true
		}
	}
	return false
}

// AttackedWithSubtypeThisTurn is how many creatures playerID declared as
// attackers this turn that had `subtype` as they attacked (a changeling
// counts for any creature type).
//
// Caller must hold g.mu.
func (g *Game) AttackedWithSubtypeThisTurn(playerID uuid.UUID, subtype string) int {
	if playerID == uuid.Nil || subtype == "" {
		return 0
	}
	return subtypeTallyCount(g.TurnTally.AttackedSubtypes, playerID, subtype)
}

// DistinctAttackersThisTurn is how many different creature OBJECTS
// playerID has declared as attackers this turn: a creature that attacks
// in two combat phases counts once (Boros Strike-Captain's "three or
// more creatures" is three creatures, not three declarations). Only the
// active player declares attackers (CR 508.1).
//
// Caller must hold g.mu.
func (g *Game) DistinctAttackersThisTurn(playerID uuid.UUID) int {
	if playerID == uuid.Nil || playerID != g.activePlayerIDLocked() {
		return 0
	}
	seen := map[PermissionCardRef]bool{}
	for _, a := range g.TurnTally.Attacks {
		seen[PermissionCardRef{ID: a.Attacker, Epoch: a.Epoch}] = true
	}
	return len(seen)
}

// AttackedWithObjectThisTurn reports whether playerID declared the
// object (id, epoch) as an attacker this turn.
//
// Caller must hold g.mu.
func (g *Game) AttackedWithObjectThisTurn(playerID, id uuid.UUID, epoch int) bool {
	if playerID == uuid.Nil || id == uuid.Nil || playerID != g.activePlayerIDLocked() {
		return false
	}
	for _, a := range g.TurnTally.Attacks {
		if a.Attacker == id && a.Epoch == epoch {
			return true
		}
	}
	return false
}

// LoreCounterPutOnThisTurn reports whether a lore counter was put on the
// Saga object (id, epoch) this turn, whoever put it there.
//
// Caller must hold g.mu.
func (g *Game) LoreCounterPutOnThisTurn(id uuid.UUID, epoch int) bool {
	for _, r := range g.TurnTally.LoreCounterPut {
		if r.ID == id && r.Epoch == epoch {
			return true
		}
	}
	return false
}
