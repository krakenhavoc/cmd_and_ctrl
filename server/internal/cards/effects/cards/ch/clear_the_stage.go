package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clear the Stage — Instant {4}{B} (Ravnica Allegiance):
//
//	"Target creature gets -3/-3 until end of turn. If you control a creature with
//	 power 4 or greater, you may return up to one target creature card from your
//	 graveyard to your hand."
//
// Two target clauses: the creature, and up to one creature card in your graveyard.
// The power-4 check is made as the spell resolves, after the -3/-3 (CR 608.2h), so
// a creature the shrink killed is not counted. The returning is an "up to one"
// target chosen as the spell is cast; the card is returned when the condition holds.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77b6a3db-c098-4b06-bc7f-136eef892d30",
		Name:         "Clear the Stage",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
		),
		Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, t.ID, nil, "Clear the Stage — -3/-3", game.ModifyPTMod(-3, -3))
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.CountYouControl(ctx, effects.And(effects.Creature(), effects.PowerGE(4))) == 0 {
					return nil
				}
				t, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}),
		),
	})
}
