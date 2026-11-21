package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "334f673e-7cfa-432c-93aa-41adfbb99113",
		Name:         "Lavabelly Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: lavabellySliverGrant,
			Triggered: []game.TriggeredAbility{
				TriggerWithPurpose(Targeting(WhenThisEnters("Lavabelly Sliver — 1 damage to target player or planeswalker and you gain 1 life", gapB106OneDamageAndOneLife), targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 1))),
			},
			Text: "When this creature enters, it deals 1 damage to target player or planeswalker and you gain 1 life.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, lavabellySliverGrant),
		},
	})
}
