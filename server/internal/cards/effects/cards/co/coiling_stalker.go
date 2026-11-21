package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coiling Stalker — Creature — Snake Ninja {1}{G}, 2/1 (Neon Dynasty):
//
//	"Ninjutsu {1}{G}
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on target creature you control that doesn't have a +1/+1
//	 counter on it."
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b), so a creature that gained a counter in response is
// no longer a legal target and gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54dc1c02-2d71-465b-a285-d6d11113991c",
		Name:         "Coiling Stalker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Coiling Stalker — put a +1/+1 counter on target creature you control that doesn't have a +1/+1 counter on it",
				effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control that doesn't have a +1/+1 counter on it", effects.YouControl(), effects.B13WithoutPlusOneCounter())),
		},
	})
}
