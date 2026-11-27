package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Council of Orzhova — Legendary Creature — Spirit {W}{W}{B}{B}, 4/4
// (Guildpact):
//
//	"When Ghost Council of Orzhova enters, target opponent loses 1 life and
//	 you gain 1 life.
//	 {1}, Sacrifice a creature: Exile Ghost Council of Orzhova. Return it to
//	 the battlefield under its owner's control at the beginning of the next
//	 end step."
//
// The drain is Blind Hunter's with an opponent as the target. The second
// ability is Fleeting Spirit's blink with a creature sacrifice in the cost:
// the Council itself is a legal sacrifice (Carrion Feeder's shape), in which
// case there is nothing left to exile and nothing returns. It comes back as a
// new object at the next end step (CR 400.7, CR 603.7), entering again, so
// the drain happens again.
//
// No simplification.
func init() {
	const name = "Ghost Council of Orzhova"
	effects.Register(effects.Spec{
		OracleID:     "2ba16e4d-2de7-4906-801a-4ba2b90b5050",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(name+" — target opponent loses 1 life and you gain 1 life",
					effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1})),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Exile Ghost Council of Orzhova. Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileThisThenScheduleReturn(effects.NewContext(g, item), item, name+" — return Ghost Council of Orzhova")
			},
		}},
	})
}
