package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scattershot Archer — Creature — Elf Archer {G}, 1/2 (Conflux, #5587):
//
//	"{T}: This creature deals 1 damage to each creature with flying."
//
// Squall's sweep on a {T} row: each creature with flying as the ability
// resolves, anyone's (the Archer too, if something gave it flying), read
// after the layers. The Archer is the source of the damage (CR 120.3),
// as it last existed if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a38ef622-d793-4867-a818-5c27982608b7",
		Name:         "Scattershot Archer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: This creature deals 1 damage to each creature with flying.",
			// ADR 0126 §6: only creatures with flying.
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.HasKeyword("flying")), 1)
			},
		}},
	})
}
