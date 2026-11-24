package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wyll's Reversal — Instant {2}{R} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP2 b05 #5948):
//
//	"Choose target spell or ability with one or more targets. Roll a d20
//	 and add the greatest power among creatures you control.
//	 1—14 | You may choose new targets for that spell or ability.
//	 15+ | You may choose new targets for that spell or ability. Then copy
//	 it. You may choose new targets for the copy."
//
// The target clause admits a spell or an activated or triggered ability
// with at least one target slot (ItemHasATarget). The greatest power is
// read as the spell resolves and rides the roll as its modifier (CR 706.2),
// so the table reads the sum (CR 706.3a) while the die stays natural. Both
// rows open the optional choose-new-targets offer (CR 115.7, Deflecting
// Swat's shape). On 15+ the copy waits for that answer (a ThenSteps walk,
// CR 608.2c) and is made from the item as it then stands: a spell copy or an
// ability copy (CR 707.10), the caster's, with its own "you may choose new
// targets" (CR 707.10c). A target that left the stack before the tail ran
// does nothing; a sum below 1 is no row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d35cef8-a52d-45fb-8f5f-cccea26826d0",
		Name:         "Wyll's Reversal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpellOrAbility("target spell or ability with one or more targets", effects.ItemHasATarget()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			return effects.RollDice{
				Sides: 20, N: 1,
				Modifier: effects.GreatestPowerAmongCreaturesYouControl()(game.CostQuery{Game: ctx.Game, Controller: ctx.Controller()}),
				Then:     wyllsReversalRolled,
			}.Apply(ctx)
		},
	})
}

// wyllsReversalRolled reads the table on the sum. An on-disk key.
var wyllsReversalRolled = effects.CardTail("dice/wylls-reversal-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	sum := effects.DiceResult(o)
	switch {
	case len(o.Rolls) == 0 || sum < 1:
		return nil
	case sum >= 15:
		return wyllsReversalRetargetThenCopy.Run(ctx)
	}
	return effects.Gap2B05MayChooseNewTargetsForTheTarget(ctx)
})

// wyllsReversalRetargetThenCopy is the 15+ row: the offer, then the copy
// once it is answered. An on-disk key.
var wyllsReversalRetargetThenCopy = effects.RegisterThenSteps("steps/wylls-reversal-retarget-then-copy",
	effects.Instruction(effects.Gap2B05MayChooseNewTargetsForTheTarget),
	effects.Instruction(effects.Gap2B05CopyTheTargetedSpellOrAbility),
)
