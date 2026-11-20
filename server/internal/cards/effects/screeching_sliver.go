package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "88485575-e491-4ab7-a020-e52dd92ae085",
		Name:         "Screeching Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: screechingSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Target player mills a card",
				Cost:    TapCost(),
				Targets: TargetPlayer("target player"),
				Effect:  TargetPlayerMills(1),
			}},
			Text: "{T}: Target player mills a card.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, screechingSliverGrant),
		},
	})
}
