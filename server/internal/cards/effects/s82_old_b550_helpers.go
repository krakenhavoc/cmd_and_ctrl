package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b550_helpers.go — Lorwyn, enters and dies triggers (S82 older sets,
// batch 550, #5593): the shared shapes of Battlewand Oak and Ashling's
// Prerogative.

// S82OldB550ThisGetsPlusTwo is "this creature gets +2/+2 until end of turn" for
// a triggered ability. The creature must still be the permanent that triggered
// (CR 400.7).
func S82OldB550ThisGetsPlusTwo(name string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: name + " — +2/+2 until end of turn"}.Apply(NewContext(g, item))
	}
}

// s82OldB550HasParity reports whether mana value mv is odd (odd) or even
// (zero is even, CR 202.3).
func s82OldB550HasParity(mv int, odd bool) bool { return (mv%2 != 0) == odd }

// S82OldB550ManaValueParity is the AppliesTo of "each creature with mana value
// of the chosen quality": every creature, any player's, whose mana value is
// odd (odd) or even. A card whose cost can't be read matches neither.
func S82OldB550ManaValueParity(odd bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, _ *game.Card) bool {
		if !target.IsCreature() {
			return false
		}
		mv, ok := target.ParsedManaValue()
		return ok && s82OldB550HasParity(mv, odd)
	}
}

// S82OldB550EntersTappedUnlessParity is "each creature without mana value of
// the chosen quality enters tapped": a CR 614 replacement on the move of every
// creature whose mana value is not odd (odd) or not even.
func S82OldB550EntersTappedUnlessParity(odd bool, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || !entering.IsCreature() {
				return false
			}
			mv, ok := g.ManaValueForEffect(entering)
			return ok && !s82OldB550HasParity(mv, odd)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
