package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Omens — Sorcery {4}{B} (Secrets of Strixhaven):
//
//	"Converge — Target player discards X cards, where X is the number
//	 of colors of mana spent to cast this spell."
//
// Converge is an ability word and the colours spent are read back at
// resolution from the stack item, as Painful Truths does:
// WantsDistinctColors pays the generic half with colours not yet spent and
// ReadsManaSpent makes a waived payment ask what was paid. The target
// player picks the discards (CR 701.9a); a hand smaller than X gives up
// what it has.
func init() {
	Register(Spec{
		OracleID:            "863cbd12-da20-456e-8698-06fba4d9bd5c",
		Name:                "Arcane Omens",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Targets:             TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			n := ctx.ColorsSpentCount()
			if n <= 0 {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: item.Targets[0].ID,
				Source: item.SourceCardID,
				N:      n,
			})
			return nil
		},
	})
}
