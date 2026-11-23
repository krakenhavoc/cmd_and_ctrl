package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kill! Maim! Burn! — Instant {3}{B}{R} (Warhammer 40,000 Commander):
//
//	"Choose one or more —
//	 • Destroy target artifact.
//	 • Destroy target creature.
//	 • Kill! Maim! Burn! deals 3 damage to target player."
//
// Each chosen bullet announces its own target (CR 700.2) and runs in
// printed order; a bullet whose target has gone is skipped on its own
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05c7d88d-c081-411b-a530-969b5816d792",
		Name:         "Kill! Maim! Burn!",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), killMaimBurnDestroy),
			effects.ModeDoing("Destroy target creature.", effects.TargetCreature("target creature"), killMaimBurnDestroy),
			effects.ModeDoing("Kill! Maim! Burn! deals 3 damage to target player.", effects.TargetPlayer("target player"), killMaimBurnDamage),
		),
	})
}

func killMaimBurnDestroy(_ *game.StackItem, ctx *effects.Context, occ int) error {
	if t, ok := effects.ModeTarget(ctx, occ); ok && t.Kind == game.TargetCard {
		return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

func killMaimBurnDamage(_ *game.StackItem, ctx *effects.Context, occ int) error {
	if t, ok := effects.ModeTarget(ctx, occ); ok && t.Kind == game.TargetPlayer {
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
	}
	return nil
}
