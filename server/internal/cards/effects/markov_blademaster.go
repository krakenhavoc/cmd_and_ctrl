package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Markov Blademaster — Creature — Vampire Warrior {1}{R}{R}, 1/1:
//
//	"Double strike
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// It deals first-strike and regular damage, so the trigger fires up to
// twice a combat, once per damage step that hits a player. A Blademaster
// that has left by resolution gets nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "774e474a-f2a1-4b62-8cfe-40ad509f9062",
		Name:            "Markov Blademaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Markov Blademaster — put a +1/+1 counter on it",
				Do(Instruction(func(ctx *Context) error {
					return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}))),
		},
	})
}
