package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Repossession — Sorcery {B} (Dominaria United):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this
//	 spell.)
//	 Return target creature card from your graveyard to your hand. You
//	 gain 2 life. If this spell was kicked, return another target
//	 permanent card from your graveyard to your hand."
//
// Kicked, the target statement is two clauses, the second Distinct
// ("another"), chosen only when the kicker is paid (Falling Timber's
// shape, CR 601.2b–c). Printed order is resolution order (CR 608.2c):
// the creature card, the life, then the other card. A card that left the
// graveyard in response is skipped and the rest still happens
// (CR 608.2b); if every target is gone the spell does nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c0eeb01-62e7-4a7f-923d-a7b830e4c83b",
		Name:         "Urborg Repossession",
		Purpose:      game.Purpose{LifeGain: 2},
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{1}{G}"),
			effects.Clauses(
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.Distinct(effects.TargetCardInGraveyard("another target permanent card from your graveyard", effects.YouOwn(), effects.Permanent())),
			))},
		Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return urborgReturn(ctx, 0)
			}),
			effects.GainLife{Amount: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				return urborgReturn(ctx, 1)
			}),
		),
	})
}

// urborgReturn returns the card named by clause `slot` from the graveyard
// to its owner's hand, when it is still a legal target.
func urborgReturn(ctx *effects.Context, slot int) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
}
