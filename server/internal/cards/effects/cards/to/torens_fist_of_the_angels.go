package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torens, Fist of the Angels — Legendary Creature — Human Cleric {1}{G}{W},
// 2/2 (Crimson Vow):
//
//	"Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 Whenever you cast a creature spell, create a 1/1 green and white Human
//	 Soldier creature token with training."
//
// Training is the shared Training() trigger. The cast trigger resolves ahead
// of the spell; the token carries its own training as a catalog token template
// (HumanSoldierTrainingToken).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cb96408-9195-4e41-ad9a-f8c74fbad083",
		Name:         "Torens, Fist of the Angels",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Training("Torens, Fist of the Angels — training: put a +1/+1 counter on it"),
			effects.WheneverYouCast(effects.Creature(), "Torens, Fist of the Angels — create a 1/1 green and white Human Soldier creature token with training",
				effects.Do(effects.CreateToken{Template: effects.HumanSoldierTrainingToken(), N: 1})),
		},
	})
}
