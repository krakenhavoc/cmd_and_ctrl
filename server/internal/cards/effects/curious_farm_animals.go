package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curious Farm Animals — Creature — Boar Elk Bird Ox {W}, 1/1 (Avatar:
// The Last Airbender, #4459):
//
//	"When this creature dies, you gain 3 life.
//	 {2}, Sacrifice this creature: Destroy up to one target artifact or
//	 enchantment."
//
// The sacrifice is a cost, so activating it also makes the dies trigger:
// the life comes either way. "Up to one" lets the ability be activated
// with no target at all (a sacrifice for the life alone); a chosen
// target that has gone by resolution is skipped (destroyFirstLegalTarget).
// Smart autopass reads that targetless announcement as a sacrifice outlet.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e451b5aa-5832-42b3-8d72-f67e126a529f",
		Name:         "Curious Farm Animals",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Curious Farm Animals — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, Sacrifice this creature: Destroy up to one target artifact or enchantment.",
			Cost:  Plus(ManaCost("{2}"), SacrificeThis()),
			// ADR 0142: announced with no target, the row is a sacrifice of
			// a creature at will (for its dies trigger's life).
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Targets: TargetPermanent("up to one target artifact or enchantment", Or(Artifact(), Enchantment())).WithCount(0, 1),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
