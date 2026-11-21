package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emeria Captain — Creature — Angel Warrior {3}{W}, 1/1 (Zendikar Rising,
// #5508):
//
//	"Flying, vigilance
//	 When this creature enters, put a +1/+1 counter on it for each
//	 creature in your party. (Your party consists of up to one each of
//	 Cleric, Rogue, Warrior, and Wizard.)"
//
// Flying and vigilance are printed. The party (CR 700.8, finB9PartySize)
// is counted as the trigger resolves (CR 700.8a), the Captain itself
// being the Warrior if nothing else is. The counters go on the Captain
// through CounterOnThis, so a Captain that left in response gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "151cafcb-a589-4681-a914-7313246a7626",
		Name:            "Emeria Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Emeria Captain — put a +1/+1 counter on it for each creature in your party",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: effects.FinB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
				}))),
		},
	})
}
