package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lavabelly Sliver — Creature — Sliver {1}{R}{W}, 2/2 (Commander Masters):
//
//	"Sliver creatures you control have "When this creature enters, it deals
//	 1 damage to target player or planeswalker and you gain 1 life.""
//
// A granted enters trigger (ADR 0093) on your own Slivers, this one
// included. Balefire Liege's player-or-planeswalker target, then the life
// gain.
//
// No simplification.
const lavabellySliverGrant = "lavabelly-sliver/enters"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "334f673e-7cfa-432c-93aa-41adfbb99113",
		Name:         "Lavabelly Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: lavabellySliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Lavabelly Sliver — 1 damage to target player or planeswalker and you gain 1 life", effects.GapB106OneDamageAndOneLife), effects.TargetPlayerOrPlaneswalker()), effects.ForTargets(effects.DamageToTarget(0, 1))),
			},
			Text: "When this creature enters, it deals 1 damage to target player or planeswalker and you gain 1 life.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, lavabellySliverGrant),
		},
	})
}
