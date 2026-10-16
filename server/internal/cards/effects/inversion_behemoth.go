package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inversion Behemoth — Creature — Eldrazi {2}{C}{C}, 2/9:
//
//	"At the beginning of combat on your turn, switch the power and
//	 toughness of each of any number of target creatures until end of
//	 turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps. "Any number" includes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fedcb737-723b-4a38-b86c-dad6ac7bec7b",
		Name:         "Inversion Behemoth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
				"Inversion Behemoth — switch the power and toughness of any number of target creatures",
				func(g *game.Game, item *game.StackItem) error {
					return switchEachTargetPT(NewContext(g, item), "Inversion Behemoth — switch power and toughness")
				}), TargetCreature("any number of target creatures").WithCount(0, 0)),
		},
	})
}
