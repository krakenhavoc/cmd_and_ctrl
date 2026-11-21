package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kird Chieftain — Creature — Ape {3}{R}, 3/3 (Magic 2015, #5556):
//
//	"This creature gets +1/+1 as long as you control a Forest.
//	 {4}{G}: Target creature gets +2/+2 and gains trample until end of
//	 turn."
//
// Kird Ape's Forest-gated pump, and Ironhoof Boar's channel pump
// (neoB07TargetGetsUntilEOT) as an activated ability on any creature,
// the target re-checked at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a7cca31-ccca-459b-9ac6-c8f384136ba5",
		Name:         "Kird Chieftain",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Forest.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Forest")),
			Apply:     effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: Target creature gets +2/+2 and gains trample until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
			Cost:    effects.ManaCost("{4}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Kird Chieftain", 2, 2, "trample"),
		}},
	})
}
