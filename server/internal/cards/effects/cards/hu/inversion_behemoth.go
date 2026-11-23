package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fedcb737-723b-4a38-b86c-dad6ac7bec7b",
		Name:         "Inversion Behemoth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Inversion Behemoth — switch the power and toughness of any number of target creatures",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SwitchEachTargetPT(effects.NewContext(g, item), "Inversion Behemoth — switch power and toughness")
				}), effects.TargetCreature("any number of target creatures").WithCount(0, 0)),
		},
	})
}
