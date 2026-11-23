package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harmonic Sliver — Creature — Sliver {1}{G}{W}, 1/1 (Time Spiral, #5596):
//
//	"All Slivers have "When this permanent enters, destroy target
//	 artifact or enchantment.""
//
// A granted enters trigger (ADR 0093, CR 603.6a) on every Sliver at the
// table. Each Sliver that enters while a Harmonic Sliver is out, Harmonic
// Sliver itself included, destroys a target artifact or enchantment. It is
// not a "may": with only its controller's own artifacts or enchantments to
// target, one of them is destroyed, and with none at all the trigger is
// removed.
//
// No simplification.
const harmonicSliverGrant = "harmonic-sliver/destroy"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c06c093-18ab-49e6-a82d-38e826d25016",
		Name:         "Harmonic Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: harmonicSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.WhenThisEnters("Harmonic Sliver — destroy target artifact or enchantment", effects.DestroyChosenPermanent),
					effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
			},
			Text: "When this permanent enters, destroy target artifact or enchantment.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, harmonicSliverGrant),
		},
	})
}
