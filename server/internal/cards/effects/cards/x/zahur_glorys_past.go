package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zahur, Glory's Past — Legendary Creature — Zombie Cat Warrior {W}{B},
// 3/2:
//
//	"Start your engines!
//	 Sacrifice another creature: Surveil 1. Activate only once each turn.
//	 Max speed — Whenever a nontoken creature you control dies, create a
//	 tapped 2/2 black Zombie creature token."
//
// ADR 0138 (#2122, #2711). The surveil is an ordinary sacrifice outlet
// limited by the activation tally (CR 602.1b). The max speed trigger
// includes Zahur himself dying: it is read from the owner's speed when
// the creature is no longer on the battlefield (CR 702.178b), and the
// token enters tapped through the template.
//
// No simplification.
func init() {
	const surveilLabel = "Sacrifice another creature: Surveil 1. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "fb82cd81-0e71-43ab-88f9-b1fa158f5138",
		Name:            "Zahur, Glory's Past",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Activated: []effects.ActivatedAbility{{
			Label:     surveilLabel,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Condition: effects.OncePerTurnActivation(surveilLabel),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Surveil{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.MaxSpeedTrigger(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == effects.SpeedPlayerOf(source)
			}, "Zahur, Glory's Past — create a tapped 2/2 black Zombie creature token",
				func(g *game.Game, item *game.StackItem) error {
					tmpl := effects.BlackZombieToken()
					tmpl.Tapped = true
					return effects.CreateToken{Template: tmpl, N: 1}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
