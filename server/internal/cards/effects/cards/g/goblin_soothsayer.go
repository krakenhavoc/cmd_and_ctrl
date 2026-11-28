package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Soothsayer — Creature — Goblin Shaman {R}, 1/1 (Mirage, S82 OLD b797,
// #5633):
//
//	"{R}, {T}, Sacrifice a Goblin: Red creatures get +1/+1 until end of turn."
//
// Facevaulter's cost: the Goblin is a cost, so the Soothsayer may sacrifice
// itself. The red creatures are those on the battlefield as it resolves
// (CR 611.2c), judged by colour then; later arrivals and recolours are not
// boosted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9277df9-1417-45bc-bebf-d9a7595f54c9",
		Name:         "Goblin Soothsayer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice a Goblin: Red creatures get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.B25SacrificeAGoblin()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.OfColor("R")), Power: 1, Toughness: 1,
					Label: "Goblin Soothsayer — red creatures get +1/+1 until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
