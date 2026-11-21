package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonehorn Dignitary — Creature — Rhino Soldier {3}{W}, 1/4 (Magic 2012,
// #5573):
//
//	"When this creature enters, target opponent skips their next
//	 combat phase."
//
// An enters trigger targeting an opponent, who skips their next combat
// phase: Blinding Angel's record (game.SkipNextCombatPhaseForEffect, CR
// 500.11, 614.10), written for the targeted player. A target that left the
// game in response is no longer legal and skips nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65368569-68bf-4498-9979-d2935942ca49",
		Name:         "Stonehorn Dignitary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Stonehorn Dignitary — target opponent skips their next combat phase", stonehornDignitarySkips),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// stonehornDignitarySkips records one skipped combat phase for each
// player target still legal.
func stonehornDignitarySkips(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			g.SkipNextCombatPhaseForEffect(t.ID, ctx.Source(), item.Label)
		}
	}
	return nil
}
