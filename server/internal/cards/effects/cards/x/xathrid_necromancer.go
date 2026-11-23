package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xathrid Necromancer — Creature — Human Wizard {2}{B}, 2/2 (Magic 2014, #5563):
//
//	"Whenever this creature or another Human creature you control dies, create a tapped 2/2 black Zombie creature token."
//
// "This creature or another Human creature you control": the Necromancer's
// own death counts whatever its types, and another creature counts when it
// was a Human under its controller's control as it last existed (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6b38874-c8be-47f7-a656-ba49ec9e967f",
		Name:         "Xathrid Necromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, effects.M14ThisOrAnotherHumanYouControlDied,
				"Xathrid Necromancer — create a tapped 2/2 black Zombie creature token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/2 black Zombie")).EntersTapped(), N: 1})),
				game.Purpose{Tokens: 1, CreatureTokens: 1}),
		},
	})
}
