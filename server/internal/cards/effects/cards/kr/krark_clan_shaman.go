package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krark-Clan Shaman — Creature — Goblin Shaman {R}, 1/1 (Mirrodin, #5606):
//
//	"Sacrifice an artifact: This creature deals 1 damage to each creature
//	 without flying."
//
// Barrage Ogre's artifact sacrifice with Rolling Temblor's sweep over the
// creatures without flying, read through the layers as the ability resolves.
// The Shaman may sacrifice itself if it is an artifact; it then deals the
// damage as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "198d1792-d638-4760-a476-7ec3af495610",
		Name:         "Krark-Clan Shaman",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature deals 1 damage to each creature without flying.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.SacrificeN(1, "an artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
			},
		}},
	})
}
