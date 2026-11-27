package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar Golem — {7} Artifact Creature — Golem, */*:
//
//	"Trample
//	 Altar Golem's power and toughness are each equal to the number of creatures
//	 on the battlefield.
//	 This creature doesn't untap during your untap step.
//	 Tap five untapped creatures you control: Untap this creature."
//
// The power and toughness are a characteristic-defining ability (layer 7a)
// counting every creature, the Golem included. Tapping creatures is not the {T}
// symbol, so a creature that arrived this turn may help, and the Golem may be
// one of the five.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "ac143b8e-f993-4989-8833-e26dc9faffa4",
		Name:                  "Altar Golem",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"trample"},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of creatures on the battlefield.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.S82OldB523CreaturesOnTheBattlefield(g)
				c.Power = n
				c.Toughness = n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap five untapped creatures you control: Untap this creature.",
			Cost:    effects.S82OldB523TapFiveUntappedCreatures(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.B31UntapSelf,
		}},
	})
}
