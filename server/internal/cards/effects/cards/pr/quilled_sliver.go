package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quilled Sliver — Creature — Sliver {1}{W}, 1/1 (Time Spiral, #5596):
//
//	"All Slivers have "{T}: This permanent deals 1 damage to target
//	 attacking or blocking creature.""
//
// A granted tap ability (ADR 0093) on every Sliver at the table. The
// target must be an attacking or blocking creature, at announce and again
// at resolution (CR 608.2b); the Sliver that has it is the damage's
// source.
//
// No simplification.
const quilledSliverGrant = "quilled-sliver/ping"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "beb39342-5272-41a4-a470-9f2bf0cd0f92",
		Name:         "Quilled Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: quilledSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This permanent deals 1 damage to target attacking or blocking creature",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
				Effect:  effects.SourceDealsOneToFirstTarget,
			}},
			Text: "{T}: This permanent deals 1 damage to target attacking or blocking creature.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, quilledSliverGrant),
		},
	})
}
