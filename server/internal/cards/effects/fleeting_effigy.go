package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleeting Effigy — Creature — Elemental {R}, 2/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Haste
//	 At the beginning of your end step, return this creature to its
//	 owner's hand. (Return it only if it's on the battlefield.)
//	 {2}{R}: This creature gets +2/+0 until end of turn."
//
// The end-step trigger returns the Effigy only while it is still the
// permanent the ability came from (CR 400.7): one that died, or left and
// came back, in response stays where it is. The pump is Freejam Regent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d6d9b5d2-0ab6-4637-a468-5149b7c7d0cf",
		Name:            "Fleeting Effigy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Fleeting Effigy — return this creature to its owner's hand",
				Do(Instruction(func(ctx *Context) error {
					if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return BounceToHand{Target: ctx.Item.SourceCardID}.Apply(ctx)
				}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{R}"),
			Effect:  thisGetsUntilEndOfTurn(2, 0, "Fleeting Effigy — +2/+0"),
		}},
	})
}
