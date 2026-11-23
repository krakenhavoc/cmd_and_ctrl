package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2bc3d5b2-076f-4e01-87be-28d7eb70a2db",
		Name:         "Dormant Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: dormantSliverGrant,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisEnters("Dormant Sliver — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "When this permanent enters, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "defender"),
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, dormantSliverGrant),
		},
	})
}
