package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sinking Feeling — Enchantment — Aura {2}{U} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 Enchanted creature has "{1}, Put a -1/-1 counter on this creature: Untap this
//	 creature.""
//
// Bitter Chill's untap restriction, and Presence of Gond's granted activated
// ability with Devoted Druid's untap-for-a-counter body, paid by the creature's
// controller with the mana.
//
// No simplification.
func init() {
	const grant = "sinking-feeling/untap"
	const text = "{1}, Put a -1/-1 counter on this creature: Untap this creature."
	effects.Register(effects.Spec{
		OracleID:              "8a5e925f-320d-4bb7-9374-0384e7ac41e7",
		Name:                  "Sinking Feeling",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.AddCounterToThis(game.CounterMinusOne, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
