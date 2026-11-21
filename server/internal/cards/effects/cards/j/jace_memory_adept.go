package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace, Memory Adept — Legendary Planeswalker — Jace {3}{U}{U}, loyalty 4
// (Magic 2012, #5573):
//
//	"+1: Draw a card. Target player mills a card.
//	 0: Target player mills ten cards.
//	 −7: Any number of target players each draw twenty cards."
//
// Three loyalty rows. The +1 draws, then the target player mills one
// (Jace, the Perfected Mind's mill); the 0 mills ten; the -7 has each
// player target still legal draw twenty, in the order they were picked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a4dc133-541f-456b-a95f-3d707e1c95e9",
		Name:            "Jace, Memory Adept",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card. Target player mills a card.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayer("target player"),
				Effect: effects.Do(
					effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error { return effects.JacePerfectedMindMill(ctx, 1) }),
				),
			},
			{
				Label:   "0: Target player mills ten cards.",
				Cost:    effects.LoyaltyCost(0),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.B31MillChosenPlayer(10),
			},
			{
				Label:   "−7: Any number of target players each draw twenty cards.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetPlayer("any number of target players").WithCount(0, 0),
				Effect:  effects.Do(effects.Instruction(jaceMemoryAdeptEachDrawsTwenty)),
			},
		},
	})
}

// jaceMemoryAdeptEachDrawsTwenty is the -7: each player target still
// legal draws twenty cards.
func jaceMemoryAdeptEachDrawsTwenty(ctx *effects.Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (effects.DrawCards{Player: t.ID, N: 20}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
