package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jetmir's Fixer — Creature — Cat Warrior {R}{G}, 2/2 (Streets of New
// Capenna):
//
//	"{R}{G}: This creature gets +1/+1 until end of turn. If mana from a
//	 Treasure was spent to activate this ability, put a +1/+1 counter on
//	 this creature instead."
//
// The payment of the activation is read as it resolves
// (StackItem.Paid, #3855): the Fixer declares ReadsOwnActivations with
// ReadsSourceKinds, so a payment waived outside the pool asks which mana
// an artifact made. A payment the engine did not see counts as no
// Treasure mana: weaker, never stronger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f87df7d4-99e2-4340-9d0a-b96c397555db",
		Name:           "Jetmir's Fixer",
		ReadsManaSpent: game.ReadsOwnActivations | game.ReadsSourceKinds,
		Completeness:   effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{G}: This creature gets +1/+1 until end of turn. If mana from a Treasure was spent to activate this ability, put a +1/+1 counter on this creature instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}{G}"),
			Effect: effects.S82SncB11PumpOrCounter("Jetmir's Fixer — +1/+1 until end of turn",
				func(ctx *effects.Context) bool { return ctx.ManaSpent().FromTreasure() }),
		}},
	})
}
