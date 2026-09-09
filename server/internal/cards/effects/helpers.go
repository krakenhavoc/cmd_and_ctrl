package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cardDied reports whether an EventLTB marks `source` going to the
// graveyard from the battlefield — i.e. it "died" (CR 700.4) — as
// opposed to being exiled, bounced, or tucked into the library.
// Dies-triggers gate their AppliesTo on this: a plain EventLTB fires
// for every battlefield exit, so matching on the event kind alone
// would mis-fire a "when ~ dies" ability on a bounce or a Path to
// Exile. The harvester stamps EventLTB.NewZone at every exit site.
// Added in S19 sub-PR 4.
func cardDied(ev game.Event, source *game.Card) bool {
	return ev.CardID == source.InstanceID && ev.NewZone == game.ZoneGraveyard
}

// IsBasicLand reports whether a card's type line contains the
// "basic land" supertype (case-insensitive substring). Used by
// tutor / fetch primitives that need to match Forest / Island /
// Mountain / Plains / Swamp / Wastes without enumerating each
// subtype. Path to Exile, Cultivate, and Solemn Simulacrum share
// this predicate; keep it here rather than duplicating the
// lowercase loop per-card.
func IsBasicLand(c game.Card) bool {
	return containsFoldASCII(c.TypeLine, "basic land")
}

// containsFoldASCII is a zero-alloc case-insensitive substring
// check for ASCII-only needles. The type lines we feed this come
// straight from Scryfall and are plain ASCII; if future text goes
// Unicode the caller should switch to strings.EqualFold + slicing.
func containsFoldASCII(haystack, needle string) bool {
	if len(haystack) < len(needle) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			h := haystack[i+j]
			if h >= 'A' && h <= 'Z' {
				h += 'a' - 'A'
			}
			if h != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// combatDamageToPlayerBy reports whether ev is combat damage dealt
// to a player by a creature that `controller` controls. The
// creature is looked up live: combat damage is dealt before SBAs
// run, so an attacker that traded with its blocker is still on the
// battlefield when its damage event fires. "Whenever a creature you
// control deals combat damage to a player" (Bident of Thassa,
// Coastal Piracy) gates on this. Added in S19 sub-PR 7.
//
// Caller must hold g.mu.
func combatDamageToPlayerBy(ev game.Event, controller uuid.UUID, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.Target); p == nil {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCreature() && src.Controller == controller
}

// --- S21 sub-PR 3: aristocrats helpers ---------------------------

// diedCreature resolves the creature that just died from a dies
// (EventLTB → graveyard) event: the card as it now sits in the
// graveyard, plus ok=false when the event isn't a creature death.
//
// The card is read post-move, so Tapped / Counters are already
// cleared (CR 400.7) but TypeLine, Controller and Owner survive —
// which is what "another creature YOU CONTROL dies" needs. A
// creature that stopped being a creature before it died is a
// sandbox gap: we read the printed type line rather than the
// battlefield LKI, because the LKI map is keyed to the dying card's
// own triggers and isn't reachable from a watcher's AppliesTo.
func diedCreature(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() {
		return game.Card{}, false
	}
	return c, true
}

// IsToken reports whether a card is a token. Token type lines are
// stamped "Token Creature — Spirit" by the templates in tokens.go;
// "Token" isn't one of the supertypes ParseTypeLine knows, so a
// substring check on the printed line is the reliable test.
func IsToken(c game.Card) bool {
	return containsFoldASCII(c.TypeLine, "token")
}

// --- staples helpers (Commander staples pass) --------------------

// IsLandWithSubtype returns a SearchLibrary predicate matching any
// land whose type line carries `subtype`. Distinct from IsBasicLand:
// Nature's Lore fetches "a Forest card", which is any land with the
// Forest subtype — a Snow-Covered Forest or a Bayou both qualify,
// while IsBasicLand would admit an Island.
//
// Case-insensitive substring on the printed line, same posture as
// IsBasicLand. It does not verify the card is a land when the
// subtype is unambiguous, so callers wanting "basic" semantics
// should compose with IsBasicLand instead.
func IsLandWithSubtype(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		return containsFoldASCII(c.TypeLine, "land") && containsFoldASCII(c.TypeLine, subtype)
	}
}

// IsBasicLandExcept returns a predicate matching a basic land whose
// type line does NOT carry `subtype` — Farseek's "Plains, Island,
// Swamp, or Mountain" expressed as "a basic land that isn't a
// Forest". Cheaper and more future-proof than enumerating four
// subtypes, and it keeps the Wastes case out by accident rather
// than by omission (Wastes has no basic land type Farseek names,
// but it also isn't a Forest — a known sandbox over-permission,
// noted on the card).
func IsBasicLandExcept(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		return IsBasicLand(c) && !containsFoldASCII(c.TypeLine, subtype)
	}
}

// controllerOfTarget resolves the controller of a targeted card, for
// the "its controller …" clause on Beast Within / Generous Gift /
// Nature's Claim. Returns ok=false when the target has left the
// battlefield between announce and resolution — the caller has
// already destroyed nothing in that case, so it just skips the
// rider rather than guessing whose token it is.
func controllerOfTarget(ctx *Context, target uuid.UUID) (uuid.UUID, bool) {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return uuid.Nil, false
	}
	return c.Controller, true
}
