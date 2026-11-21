package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bumbleflower's Sharepot — Artifact {2} (Bloomburrow, #4467):
//
//	"When this artifact enters, create a Food token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	 {5}, {T}, Sacrifice this artifact: Destroy target nonland permanent.
//	 Activate only as a sorcery."
//
// Bakersbane Duo's entry Food (the catalog Food, ADR 0083) and Thaumaton
// Torpedo's sacrifice-to-destroy at sorcery speed. The target is
// re-checked at resolution (CR 608.2b). A sorcery-speed row declares no
// Answers (ADR 0142).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e152529-64b4-472c-8e50-99095af6af99",
		Name:         "Bumbleflower's Sharepot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bumbleflower's Sharepot — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, {T}, Sacrifice this artifact: Destroy target nonland permanent. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
			Effect:       effects.DestroyFirstLegalCardTarget,
		}},
	})
}
