package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladeback Sliver — Creature — Sliver {1}{R}, 2/2 (Modern Horizons):
//
//	"Hellbent — As long as you have no cards in hand, Sliver creatures you
//	 control have '{T}: This creature deals 1 damage to target player or
//	 planeswalker.'"
//
// The granted ability is a bundle (ADR 0093) handed to the Slivers its
// controller has while that controller's hand is empty; the bundle's recipient
// test is re-read on every layer pass and the static declares that it depends
// on hand sizes, so a draw, a discard or a card put into the hand changes it at
// once. The Bladeback Sliver is a Sliver creature and has the ability itself.
// The ability is the grantee's own (it is the damage source, CR 609.7).
//
// No simplification.
const bladebackSliverGrant = "bladeback-sliver/ping"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "a64656b2-0f0c-4239-9dfe-a9ddc7ada733",
		Name:         "Bladeback Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: bladebackSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to target player or planeswalker.",
		}},
		Static: []game.StaticAbility{
			effects.OldB165HellbentTribalGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, bladebackSliverGrant),
		},
	})
}
