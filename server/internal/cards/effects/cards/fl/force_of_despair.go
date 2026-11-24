package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force of Despair — Instant {1}{B}{B} (Secret Lair Drop):
//
//	"If it's not your turn, you may exile a black card from your hand rather
//	 than pay this spell's mana cost.
//	 Destroy all creatures that entered this turn."
//
// The pitch is the alternative cost of Force of Vigor and Force of Negation,
// available only off your own turn. The sweep destroys the creatures whose own
// entry was this turn (an object that entered, left and came back this turn
// counts as the new object it is), any controller's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f2d038e-6dab-4352-9ba7-ed9e063bb304",
		Name:         "Force of Despair",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		AlternativeCosts: []game.AlternativeCost{{
			Key:           "pitch",
			Label:         "If it's not your turn, exile a black card from your hand",
			ExileFromHand: effects.CardInYourHand("a black card from your hand", effects.OfColor("B")),
			PayLabel:      "a black card from your hand",
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				return !effects.IsActivePlayer(g, controller)
			},
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.S82OldB023CreatureEnteredThisTurn()}.Apply(ctx)
		},
	})
}
