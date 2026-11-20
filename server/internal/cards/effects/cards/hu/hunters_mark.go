package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "7def9bfe-6a59-4300-8134-fce7508dca38",
		Name:            "Hunter's Mark",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Purpose:         huntersMarkPurpose(),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPermanent("target creature or planeswalker you don't control",
				effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a blue permanent you don't control.",
				effects.And(effects.OfColor("U"), effects.OpponentControls())),
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: mine.ID, Power: 1, Toughness: 1, Label: "Hunter's Mark"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OneSidedBite(nil, ctx)
			}),
		),
	})
}

// huntersMarkPurpose: a one-sided fight at a creature or planeswalker, the
// first target pumped until end of turn.
func huntersMarkPurpose() game.Purpose {
	p := effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit})
	p.Answers = game.AnswerRemove
	return p
}
