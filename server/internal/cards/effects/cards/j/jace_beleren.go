package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace Beleren — Legendary Planeswalker — Jace {1}{U}{U}, starting
// loyalty 3 (Lorwyn, #5593):
//
//	"+2: Each player draws a card.
//	 −1: Target player draws a card.
//	 −10: Target player mills twenty cards."
//
// Three loyalty abilities, each activated only as a sorcery and once a turn
// (CR 606.3), the −10 only with at least ten loyalty (CR 606.6). The +2's
// draws go round the table from the controller, one card each (CR 121.2).
// The mill is twenty cards from the top, or the whole library when it has
// fewer (CR 701.17a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc2b6e58-2d6f-495e-95f3-9496105c4cba",
		Name:         "Jace Beleren",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Each player draws a card.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    effects.LoyaltyCost(2),
				Effect:  effects.EachPlayerDrawsACard,
			},
			{
				Label:   "−1: Target player draws a card.",
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetPlayer("target player"),
				Effect: effects.OnTheTargetPlayer(func(ctx *effects.Context, p uuid.UUID) error {
					return effects.DrawCards{Player: p, N: 1}.Apply(ctx)
				}),
			},
			{
				Label:   "−10: Target player mills twenty cards.",
				Cost:    effects.LoyaltyCost(-10),
				Targets: effects.TargetPlayer("target player"),
				Effect:  effects.TargetPlayerMills(20),
			},
		},
	})
}
