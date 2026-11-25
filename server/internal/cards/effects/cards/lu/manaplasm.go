package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manaplasm — Creature — Ooze {2}{G}, 1/1 (Conspiracy: Take the Crown, S82
// older sets batch 273, #5540):
//
//	"Whenever you cast a spell, this creature gets +X/+X until end of turn,
//	 where X is that spell's mana value."
//
// Jamie McCrimmon's trigger for any spell you cast: the trigger resolves
// above the spell, so X is the mana value the spell has on the stack, with its
// announced X (CR 202.3e). The pump is pinned to this Manaplasm (CR 400.7)
// and ends at cleanup.
//
// No simplification.
func init() {
	const label = "Manaplasm — gets +X/+X until end of turn, where X is that spell's mana value"
	effects.Register(effects.Spec{
		OracleID:     "9061e2ae-7385-494a-a511-0b6531d1fe0e",
		Name:         "Manaplasm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, label, func(g *game.Game, item *game.StackItem) error {
				x := effects.TriggeringSpellManaValue(g, item)
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Toughness: x, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
