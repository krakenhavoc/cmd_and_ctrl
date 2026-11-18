package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunter's Mark — Instant {3}{G} (Adventures in the Forgotten Realms):
//
//	"This spell costs {3} less to cast if it targets a blue permanent you
//	 don't control.
//	 This spell can't be countered.
//	 Target creature you control gets +1/+1 until end of turn. Then it
//	 deals damage equal to its power to target creature or planeswalker
//	 you don't control."
//
// Animist's Might's two clauses and its "costs less if it targets" seam
// (CR 601.2f), here read against the second clause: the first can only be
// a creature you control, so only a blue permanent you don't control can
// satisfy it. The pump resolves first and the damage reads the power it
// made (CR 608.2c); the damage is dealt only when both targets are still
// legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7def9bfe-6a59-4300-8134-fce7508dca38",
		Name:            "Hunter's Mark",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Purpose:         huntersMarkPurpose(),
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPermanent("target creature or planeswalker you don't control",
				Or(Creature(), Planeswalker()), OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a blue permanent you don't control.",
				And(OfColor("U"), OpponentControls())),
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: mine.ID, Power: 1, Toughness: 1, Label: "Hunter's Mark"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return oneSidedBite(nil, ctx)
			}),
		),
	})
}

// huntersMarkPurpose: a one-sided fight at a creature or planeswalker, the
// first target pumped until end of turn.
func huntersMarkPurpose() game.Purpose {
	p := ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit})
	p.Answers = game.AnswerRemove
	return p
}
