package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screeching Sliver — Creature — Sliver {U}, 1/1 (Time Spiral, #5596):
//
//	"All Slivers have "{T}: Target player mills a card.""
//
// A granted tap ability (ADR 0093) on every Sliver at the table: the
// targeted player mills one card (CR 701.17a).
//
// No simplification.
const screechingSliverGrant = "screeching-sliver/mill"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "88485575-e491-4ab7-a020-e52dd92ae085",
		Name:         "Screeching Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: screechingSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target player mills a card",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.TargetPlayerMills(1),
			}},
			Text: "{T}: Target player mills a card.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, screechingSliverGrant),
		},
	})
}
