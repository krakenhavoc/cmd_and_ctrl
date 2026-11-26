package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mark for Death — Sorcery {3}{R} (Gatecrash, S82 OLD #5565):
//
//	"Target creature an opponent controls blocks this turn if able. Untap
//	 that creature. Other creatures that player controls can't block this
//	 turn."
//
// Three instructions in printed order (CR 608.2c): Provoke's block requirement
// and untap, then a live "can't block" on the other creatures that player
// controls, sparing the target itself (game.ScopeYourCreaturesExceptObjects,
// Oriss's player scope with Last Night Together's spared objects). "Can't
// block" changes no characteristic, so the set is read as blockers are
// declared (CR 611.2c): a creature that player gets later this turn can't
// block either. The player is the target's controller as the spell resolves;
// a target gone by then makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad2e4ab0-d990-472d-8788-3f85215da51c",
		Name:         "Mark for Death",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Mark for Death — blocks this turn if able",
				}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.UntapTarget{Target: id}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				ctx.Game.OtherCreaturesOfPlayerCantBlockForEffect(ctx.Source(), c.Controller, id, effects.DurationUntilEndOfTurn(ctx),
					"Mark for Death — other creatures that player controls can't block this turn")
				return nil
			}),
		),
	})
}
