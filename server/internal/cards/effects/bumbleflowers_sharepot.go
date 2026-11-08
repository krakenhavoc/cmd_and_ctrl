package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6e152529-64b4-472c-8e50-99095af6af99",
		Name:         "Bumbleflower's Sharepot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bumbleflower's Sharepot — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}, {T}, Sacrifice this artifact: Destroy target nonland permanent. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Targets:      TargetPermanent("target nonland permanent", Not(Land())),
			Effect:       destroyFirstLegalCardTarget,
		}},
	})
}
