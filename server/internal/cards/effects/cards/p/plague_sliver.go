package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Sliver — Creature — Sliver {2}{B}{B}, 5/5 (Time Spiral, #5596):
//
//	"All Slivers have "At the beginning of your upkeep, this permanent
//	 deals 1 damage to you.""
//
// A granted triggered ability (ADR 0093) on every Sliver creature at the
// table. "Your upkeep" and "you" are the controller of the Sliver that has
// it (CR 109.5), so an opponent's Slivers hurt that opponent in their own
// upkeep, one trigger per Sliver.
//
// No simplification.
const plagueSliverGrant = "plague-sliver/upkeep"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee10dc06-26e1-4f50-a0a6-6d039e71b75f",
		Name:         "Plague Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: plagueSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Plague Sliver — this permanent deals 1 damage to you", effects.Do(effects.DamageController{N: 1})),
			},
			Text: "At the beginning of your upkeep, this permanent deals 1 damage to you.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, plagueSliverGrant),
		},
	})
}
