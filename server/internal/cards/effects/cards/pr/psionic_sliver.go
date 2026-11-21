package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psionic Sliver — Creature — Sliver {4}{U}, 2/2 (Time Spiral, #5596):
//
//	"All Sliver creatures have "{T}: This creature deals 2 damage to
//	 any target and 3 damage to itself.""
//
// A granted tap ability (ADR 0093) on every Sliver creature at the table.
// The Sliver that has it deals 2 damage to the target and then 3 to
// itself, so a Sliver with toughness 3 or less dies to its own ping as a
// state-based action (CR 704.5g). A Sliver that has left the battlefield
// by then is dealt nothing.
//
// No simplification.
const psionicSliverGrant = "psionic-sliver/ping"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea79d4ee-c7b5-4687-9e33-22f98e46e8ec",
		Name:         "Psionic Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: psionicSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 2 damage to any target and 3 damage to itself",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.TspB1PsionicPing,
			}},
			Text: "{T}: This creature deals 2 damage to any target and 3 damage to itself.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, psionicSliverGrant),
		},
	})
}
