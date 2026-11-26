package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maw of the Obzedat — Creature — Thrull {3}{W}{B}, 3/3 (Dragon's Maze, S82 OLD
// batch 404, #5564):
//
//	"Sacrifice a creature: Creatures you control get +1/+1 until end of turn."
//
// Ghave's sacrifice cost (a creature you control, itself included, chosen as the
// ability is activated and gone before it resolves, CR 602.2b) with Dauntless
// Escort's group grant: a layer-7 +1/+1 on the creatures you control as the
// ability resolves, a fixed set (CR 611.2c), so the sacrificed creature is not in
// it and a creature that arrives later is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed8ec935-3359-42a1-8ac6-cac6f0e7f905",
		Name:         "Maw of the Obzedat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "a creature", effects.Creature()),
			Effect: effects.Do(effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
				Label: "Maw of the Obzedat — creatures you control get +1/+1 until end of turn",
			}),
		}},
	})
}
