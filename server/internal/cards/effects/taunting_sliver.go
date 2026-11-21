package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taunting Sliver — Creature — Sliver {3}{U}, 3/3 (Commander Legends: Battle
// for Baldur's Gate):
//
//	"Sliver creatures you control have "When this creature enters, goad target
//	 creature an opponent controls." (Until your next turn, that creature
//	 attacks each combat if able and attacks a player other than you if
//	 able.)"
//
// A granted enters trigger (ADR 0093) on your own Slivers, this one included,
// with Disrupt Decorum's goad (GoadTarget) on the creature an opponent
// controls. The goad's controller is the controller of the trigger, the
// Sliver's controller (CR 701.15a).
//
// No simplification.
const tauntingSliverGrant = "taunting-sliver/goad"

func init() {
	Register(Spec{
		OracleID:     "fada8f16-85e8-44f1-9d2c-99da1e27851f",
		Name:         "Taunting Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: tauntingSliverGrant,
			Triggered: []game.TriggeredAbility{
				Targeting(WhenThisEnters("Taunting Sliver — goad target creature an opponent controls", GoadTarget),
					TargetCreature("target creature an opponent controls", OpponentControls())),
			},
			Text: "When this creature enters, goad target creature an opponent controls.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, tauntingSliverGrant),
		},
	})
}
