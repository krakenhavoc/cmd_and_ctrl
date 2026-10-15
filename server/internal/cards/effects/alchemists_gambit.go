package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemist's Gambit — Sorcery {1}{R}{R}:
//
//	"Cleave {4}{U}{U}{R} (You may cast this spell for its cleave cost.
//	 If you do, remove the words in square brackets.)
//	 Take an extra turn after this one. During that turn, damage can't
//	 be prevented. [At the beginning of that turn's end step, you lose
//	 the game.]
//	 Exile Alchemist's Gambit."
//
// TakeExtraTurn with DuringThatTurn (ADR 0188 Decision 3b, #3563,
// tracker #753). "During that turn" is a CR 611.2a effect created now
// whose duration starts later, so it is queued WITH the extra turn
// (game.ExtraTurn.Effects) and registered as that turn begins, lasting
// until its cleanup step. Hence:
//
//   - damage can still be prevented for the rest of the turn the spell
//     was cast in (a Fog cast in that turn prevents damage as printed);
//   - in the extra turn nothing can prevent damage, a Fog cast there
//     prevents nothing (CR 615.12), and the effect is gone at that
//     turn's cleanup;
//   - an extra turn that is skipped never begins, so the effect never
//     starts (CR 614.10a).
//
// Hard-cast, the bracketed loss is a delayed trigger bound to that turn
// (TakeExtraTurnThenLose, as in Final Fortune): it goes on the stack in
// that turn's end step, and is swept if the turn is skipped or ends
// early (the 2021-11-19 ruling). Cleaved, the brackets are removed, so
// the cast schedules no trigger at all and the caster does not lose.
// The brackets touch no target clause (there is none), hence Cleave's nil.
//
// "Exile Alchemist's Gambit" is the spell exiling itself as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "271d8db6-8606-46d7-9812-75aa22fcfd8e",
		Name:             "Alchemist's Gambit",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Cleave("{4}{U}{U}{R}", nil)},
		// ADR 0188 Decision 10: the caster's extra turn, priced by the bot.
		Purpose: game.Purpose{ExtraTurns: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			turn := TakeExtraTurn{
				DuringThatTurn: []game.Mod{{Kind: game.ModDamageCantBePrevented}},
				Label:          "Alchemist's Gambit — damage can't be prevented during that turn",
			}
			var err error
			if ctx.PaidAltCost("cleave") {
				err = turn.Apply(ctx)
			} else {
				err = TakeExtraTurnThenLose{
					Label: "Alchemist's Gambit — you lose the game",
					Turn:  turn,
				}.Apply(ctx)
			}
			if err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(ctx.Source())
		},
	})
}
