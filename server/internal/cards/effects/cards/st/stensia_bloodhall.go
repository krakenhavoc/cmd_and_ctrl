package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stensia Bloodhall — Land (Starter Commander Decks):
//
//	"{T}: Add {C}.
//	 {3}{B}{R}, {T}: This land deals 2 damage to target player or
//	 planeswalker."
//
// A colourless mana ability and an activated ability that taps the land as
// part of its cost. The damage is dealt by the land as it resolves, to the
// one target the clause allows (a creature is not a legal target), Balefire
// Liege's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8220c5fa-28dc-40d0-a38a-d8eefc2795d6",
		Name:         "Stensia Bloodhall",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{R}, {T}: This land deals 2 damage to target player or planeswalker.",
			Cost:    effects.Plus(effects.ManaCost("{3}{B}{R}"), effects.TapCost()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
