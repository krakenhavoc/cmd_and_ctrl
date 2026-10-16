package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lae'zel, Vlaakith's Champion — Legendary Creature — Gith Warrior
// {2}{W}, 3/3 (EDHREC rank 1486):
//
//	"If you would put one or more counters on a creature or
//	 planeswalker you control or on yourself, put that many plus one
//	 of each of those kinds of counters on that permanent or player
//	 instead.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// Hardened Scales for every counter kind, on creatures and
// planeswalkers alike. The counter pipeline delivers one placement
// per kind, so "plus one of each of those kinds" is +1 on each
// event; a removal (negative delta) is not a placement and is left
// alone.
//
// "OR ON YOURSELF" is live since ADR 0056 Decision 5: counters on a
// player go through the same CR 614 window as counters on a permanent
// (game.ReplacementEvent.CounterPlayer), so the experience counter and
// the energy land at one more, and the poison an opponent's infect
// creature gives Lae'zel's controller does NOT — because that is the
// opponent putting it, not "you".
//
// "If YOU would put" is narrower than Hardened Scales' passive "would
// be put", and who is putting is a fact on the event
// (game.ReplacementEvent.CounterPlacer), named by every placement since
// #3942 (game/counter_placer.go): the controller of the resolving spell
// or ability, the source's controller for a CR 120.3d damage result,
// the proliferating player for CR 701.34, the payer of a loyalty cost
// (CR 606.4), the controller of a permanent entering with counters
// (CR 122.6a). This used to fall back to the last resolver in the log,
// which outlived the resolution it named.
//
// Choose a Background is a deck-construction rule (CR 702.124), not
// an in-game effect; the deck importer's business.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c066f921-e349-45d1-8ec3-0955d10bbf19",
		Name:         "Lae'zel, Vlaakith's Champion",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
					return false
				}
				if !laezelTargetIsYours(ev, g, src.Controller) {
					return false
				}
				return ev.CounterPlacer == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Lae'zel, Vlaakith's Champion: one more counter",
		}},
	})
}

// laezelTargetIsYours is the card's "on a creature or planeswalker you
// control or on yourself": a player counter going on Lae'zel's own
// controller, or a card counter going on a creature or planeswalker
// they control.
//
// The player arm reads ev.CounterPlayer rather than looking the ID up
// as a card, which is the whole reason the two IDs are separate fields
// on the event — a player UUID handed to LookupCardForEffect finds
// nothing, which is how this clause was silently dead before ADR 0056.
func laezelTargetIsYours(ev *game.ReplacementEvent, g *game.Game, you uuid.UUID) bool {
	if ev.CounterPlayer != uuid.Nil {
		return ev.CounterPlayer == you
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok || target.Controller != you {
		return false
	}
	return target.IsCreature() || target.IsPlaneswalker()
}
