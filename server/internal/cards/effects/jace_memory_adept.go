package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "4a4dc133-541f-456b-a95f-3d707e1c95e9",
		Name:            "Jace, Memory Adept",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Draw a card. Target player mills a card.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPlayer("target player"),
				Effect: Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error { return jacePerfectedMindMill(ctx, 1) }),
				),
			},
			{
				Label:   "0: Target player mills ten cards.",
				Cost:    LoyaltyCost(0),
				Targets: TargetPlayer("target player"),
				Effect:  b31MillChosenPlayer(10),
			},
			{
				Label:   "−7: Any number of target players each draw twenty cards.",
				Cost:    LoyaltyCost(-7),
				Targets: TargetPlayer("any number of target players").WithCount(0, 0),
				Effect:  Do(Instruction(jaceMemoryAdeptEachDrawsTwenty)),
			},
		},
	})
}

// jaceMemoryAdeptEachDrawsTwenty is the -7: each player target still
// legal draws twenty cards.
func jaceMemoryAdeptEachDrawsTwenty(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (DrawCards{Player: t.ID, N: 20}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
