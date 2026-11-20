package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "151cafcb-a589-4681-a914-7313246a7626",
		Name:            "Emeria Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Emeria Captain — put a +1/+1 counter on it for each creature in your party",
				Do(Instruction(func(ctx *Context) error {
					return CounterOnThis{Kind: game.CounterPlusOne, N: finB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
				}))),
		},
	})
}
