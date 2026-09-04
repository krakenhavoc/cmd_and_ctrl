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
