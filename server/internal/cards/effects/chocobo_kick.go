package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chocobo Kick — Sorcery {1}{G} (Final Fantasy, #5358):
//
//	"Kicker—Return a land you control to its owner's hand. (You may
//	 return a land you control to its owner's hand in addition to any
//	 other costs as you cast this spell.)
//	 Target creature you control deals damage equal to its power to
//	 target creature an opponent controls. If this spell was kicked, the
//	 creature you control deals twice that much damage instead."
//
// The kicker is the return-to-hand cost on the optional slot
// (KickerReturnToHand): the land is named on cast_spell's return_ids and
// goes back to its owner's hand with the spell on the stack (CR 601.2h),
// so one tapped for this spell's mana may pay it (CR 601.2g comes first).
// With no land the spell is cast unkicked only, and it is kicked at most
// once. The bite is Bite Down's (oneSidedBiteTimes), read per slot: the
// power is the creature's as the spell resolves, and nothing is dealt if
// either target has become illegal (CR 608.2b). A copy of a kicked spell
// is kicked too (CR 707.10).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "aa00c18b-8719-4095-940a-8e35a57b46c2",
		Name:          "Chocobo Kick",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{KickerReturnToHand("a land you control", Land())},
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			times := 1
			if ctx.WasKicked() {
				times = 2
			}
			return oneSidedBiteTimes(ctx, times)
		},
	})
}
