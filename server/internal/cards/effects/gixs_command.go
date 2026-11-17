package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gix's Command — Sorcery {3}{B}{B} (The Brothers' War Commander, #5482):
//
//	"Choose two —
//	 • Put two +1/+1 counters on up to one creature. It gains lifelink
//	   until end of turn.
//	 • Destroy each creature with power 2 or less.
//	 • Return up to two creature cards from your graveyard to your hand.
//	 • Each opponent sacrifices a creature with the greatest power among
//	   creatures they control."
//
// Kayla's Command's "choose two" template (ChooseN 2..2): the bullets run
// in printed order (CR 608.2c), each waiting for the prompt the one before
// it opened (ModeSteps), so the counters are on before the destroy bullet
// reads power (the creature with two counters and power 1 survives, the
// ruling) and the graveyard pick is made before the edict. The first bullet's
// target is "up to one": with none chosen it does nothing, and lifelink
// follows the counters only while the creature is a legal target (CR
// 608.2b). The return bullet is not targeted: its two cards are chosen on
// resolution. The edict is Crackling Doom's: each opponent sacrifices one
// of the creatures with the greatest power they control, their choice among
// ties.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9270eb03-4600-40cc-83a5-ea987d5d4978",
		Name:         "Gix's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeWithPurpose(ModeSteps("Put two +1/+1 counters on up to one creature. It gains lifelink until end of turn.",
				TargetCreature("up to one target creature").WithCount(0, 1),
				Instruction(putCountersOnTheModesTargetStep(game.CounterPlusOne, 2)),
				Instruction(gixsCommandLifelink)),
				ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink"))),
			ModeWithPurpose(ModeDoing("Destroy each creature with power 2 or less.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: And(Creature(), PowerLE(2))}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
			ModeWithPurpose(ModeSteps("Return up to two creature cards from your graveyard to your hand.", nil,
				Instruction(func(ctx *Context) error {
					return ReturnChosenFromGraveyard{
						Player:   ctx.Controller(),
						Question: "Gix's Command — return up to two creature cards from your graveyard to your hand",
						Match:    func(c game.Card) bool { return c.IsCreature() },
						Max:      2,
						ToHand:   true,
					}.Apply(ctx)
				})),
				game.Purpose{Answers: game.AnswerValue}),
			ModeWithPurpose(ModeSteps("Each opponent sacrifices a creature with the greatest power among creatures they control.", nil,
				EachPlayerSacrifices{
					ExceptController: true,
					Match:            GreatestPowerYouControl(),
					Label:            "a creature with the greatest power among creatures you control",
				}),
				game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}

// gixsCommandLifelink is the first bullet's "it gains lifelink".
func gixsCommandLifelink(ctx *Context) error {
	t, ok := ModeTarget(ctx, ctx.Occurrence())
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"lifelink"},
		Label:    "Gix's Command — gains lifelink",
	}.Apply(ctx)
}
