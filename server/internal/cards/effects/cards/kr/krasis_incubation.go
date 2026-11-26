package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krasis Incubation — Enchantment — Aura {2}{G}{U} (Dragon's Maze, S82 OLD
// b405, #5564):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block, and its activated abilities
//	 can't be activated.
//	 {1}{G}{U}, Return this Aura to its owner's hand: Put two +1/+1 counters
//	 on enchanted creature."
//
// Arrest's restriction. The return is a cost (Gossamer Chains' ReturnThis), so
// "enchanted creature" is the permanent the Aura was attached to as it left
// (CR 608.2h, EnchantedCreatureAsItResolves); a creature that is gone by
// resolution gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66e2e816-5197-4151-a6c0-352c34c2eec7",
		Name:         "Krasis Incubation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{U}, Return this Aura to its owner's hand: Put two +1/+1 counters on enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}{U}"), effects.ReturnThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				host, ok := effects.EnchantedCreatureAsItResolves(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: host, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			},
		}},
	})
}
