package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shrapnel Slinger — Artifact Creature — Phyrexian Beast {1}{R}, 2/2:
//
//	"When this creature enters, you may sacrifice a creature. When you
//	 do, destroy target artifact an opponent controls."
//
// Grist's shape: the sacrifice is asked at resolution (this creature itself
// is a legal pick), and the reflexive trigger is raised only for a creature
// that really left. Its target is chosen as it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b400536c-ea8b-4237-90da-1d3ac139881d",
		Name:         "Shrapnel Slinger",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shrapnel Slinger — you may sacrifice a creature; when you do, destroy target artifact an opponent controls", func(g *game.Game, item *game.StackItem) error {
				return maySacrificeOne(g, item, "Shrapnel Slinger — sacrifice a creature to destroy target artifact an opponent controls?", Creature(), false, shrapnelSlingerSacrificed)
			}),
		},
	})
}
