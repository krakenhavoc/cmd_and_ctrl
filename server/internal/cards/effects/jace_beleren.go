package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jace Beleren — Legendary Planeswalker — Jace {1}{U}{U}, loyalty 3:
//
//	"+2: Each player draws a card.
//	 −1: Target player draws a card.
//	 −10: Target player mills twenty cards."
//
// Three one-line abilities. The +2 draws for every player still in the game
// in turn order from the controller; the other two read their player target
// as they resolve (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cc2b6e58-2d6f-495e-95f3-9496105c4cba",
		Name:            "Jace Beleren",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:  "+2: Each player draws a card.",
				Cost:   LoyaltyCost(2),
				Effect: gapB32EachPlayerDrawsACard,
			},
			{
				Label:   "−1: Target player draws a card.",
				Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
				Cost:    LoyaltyCost(-1),
				Targets: TargetPlayer("target player"),
				Effect:  theTargetPlayerDraws(1),
			},
			{
				Label:   "−10: Target player mills twenty cards.",
				Cost:    LoyaltyCost(-10),
				Targets: TargetPlayer("target player"),
				Effect:  targetPlayerMills(20),
			},
		},
	})
}
