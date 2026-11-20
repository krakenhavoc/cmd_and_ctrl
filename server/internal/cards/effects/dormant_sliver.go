package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dormant Sliver — Creature — Sliver {2}{G}{U}, 2/2 (Planar Chaos, #5595):
//
//	"All Sliver creatures have defender.
//	 All Slivers have "When this permanent enters, draw a card.""
//
// Two grants to every Sliver at the table: defender in layer 6 (CR
// 613.1f), and Harmonic Sliver's granted enters trigger, which each
// Sliver's controller draws from as that Sliver enters, this one included.
//
// No simplification.
const dormantSliverGrant = "dormant-sliver/draw"

func init() {
	Register(Spec{
		OracleID:     "2bc3d5b2-076f-4e01-87be-28d7eb70a2db",
		Name:         "Dormant Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: dormantSliverGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisEnters("Dormant Sliver — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "When this permanent enters, draw a card.",
		}},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "defender"),
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, dormantSliverGrant),
		},
	})
}
