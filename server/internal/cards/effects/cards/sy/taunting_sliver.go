package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fada8f16-85e8-44f1-9d2c-99da1e27851f",
		Name:         "Taunting Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: tauntingSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.WhenThisEnters("Taunting Sliver — goad target creature an opponent controls", effects.GoadTarget),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			},
			Text: "When this creature enters, goad target creature an opponent controls.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, tauntingSliverGrant),
		},
	})
}
