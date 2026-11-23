package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telekinetic Sliver — Creature — Sliver {2}{U}{U}, 2/2 (Time Spiral, #5596):
//
//	"All Slivers have "{T}: Tap target permanent.""
//
// A granted tap ability (ADR 0093) on every Sliver at the table. It has
// {T} in its cost, so a Sliver that came under its controller's control
// this turn can't use it unless it has haste (CR 302.6).
//
// No simplification.
const telekineticSliverGrant = "telekinetic-sliver/tap"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "31ec6a73-f742-4dc6-809d-28d3c461c2a0",
		Name:         "Telekinetic Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: telekineticSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Tap target permanent",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target permanent"),
				Effect:  effects.B36TapChosenCreature,
			}},
			Text: "{T}: Tap target permanent.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, telekineticSliverGrant),
		},
	})
}
