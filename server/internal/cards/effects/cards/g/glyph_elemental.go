package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glyph Elemental — Enchantment Creature — Elemental {1}{W}, 2/2 (Modern
// Horizons 3):
//
//	"Bestow {1}{W} (If you cast this card for its bestow cost, it's an Aura
//	 spell with enchant creature. It becomes a creature again if it's not
//	 attached.)
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 this permanent.
//	 Enchanted creature gets +1/+1 for each +1/+1 counter on this Aura."
//
// Bestow is the shared alternative cost (ADR 0141). The landfall counter
// goes on "this permanent", so it grows as a creature and as a bestowed
// Aura alike. The host's bonus is PumpAttachedPer over the counters on the
// Aura, read on every layer pass: it tracks the counters as they come, and
// when the Aura falls off it is a creature again that pumps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "288ea954-a880-4731-8fe7-93d2ff6a043f",
		Name:             "Glyph Elemental",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Glyph Elemental — put a +1/+1 counter on this permanent",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters[game.CounterPlusOne]
			}),
		},
	})
}
