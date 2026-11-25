package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eternal Thirst — Enchantment — Aura {1}{B} (Jumpstart, #5510):
//
//	"Enchant creature
//	 Enchanted creature has lifelink and "Whenever a creature an opponent
//	 controls dies, put a +1/+1 counter on this creature.""
//
// Lifelink is a layer 6 grant and the trigger is a granted bundle (ADR
// 0093), so "you" and "an opponent" are the enchanted creature's controller
// and their opponents, whoever controls the Aura. The dying creature is
// judged as it last existed (CR 603.10a): one an opponent controlled as it
// died, each death its own trigger.
//
// No simplification.
const eternalThirstGrant = "eternal-thirst/counter-on-opponents-deaths"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1b66fbf-14a3-4582-8414-53f432f96dad",
		Name:         "Eternal Thirst",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: eternalThirstGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventLTB, effects.S82OldB100OpponentsCreatureDied,
					"Eternal Thirst — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: "Whenever a creature an opponent controls dies, put a +1/+1 counter on this creature.",
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("lifelink"),
			effects.GrantAbilitiesToAttached(eternalThirstGrant),
		},
	})
}
