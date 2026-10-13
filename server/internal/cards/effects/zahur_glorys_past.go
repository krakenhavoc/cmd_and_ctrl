package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "fb82cd81-0e71-43ab-88f9-b1fa158f5138",
		Name:            "Zahur, Glory's Past",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Activated: []ActivatedAbility{{
			Label:     surveilLabel,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      SacrificeAnotherN(1, "another creature", Creature()),
			Condition: OncePerTurnActivation(surveilLabel),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			MaxSpeedTrigger(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == speedPlayerOf(source)
			}, "Zahur, Glory's Past — create a tapped 2/2 black Zombie creature token",
				func(g *game.Game, item *game.StackItem) error {
					tmpl := BlackZombieToken()
					tmpl.Tapped = true
					return CreateToken{Template: tmpl, N: 1}.Apply(NewContext(g, item))
				})),
		},
	})
}
