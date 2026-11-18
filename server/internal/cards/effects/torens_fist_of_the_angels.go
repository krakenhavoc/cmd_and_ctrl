package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2cb96408-9195-4e41-ad9a-f8c74fbad083",
		Name:         "Torens, Fist of the Angels",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			Training("Torens, Fist of the Angels — training: put a +1/+1 counter on it"),
			WheneverYouCast(Creature(), "Torens, Fist of the Angels — create a 1/1 green and white Human Soldier creature token with training",
				Do(CreateToken{Template: HumanSoldierTrainingToken(), N: 1})),
		},
	})
}
