package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Rake — Sorcery {2}{B} (Modern Horizons, #5519):
//
//	"Target player discards two cards.
//	 Overload {1}{B} (You may cast this spell for its overload cost. If you
//	 do, change "target" in its text to "each.")"
//
// Mind Rot's body, and overloaded "each player discards two cards":
// every player, the caster included, discards in APNAP order (Rankle's
// Prank's step).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d8e0d40-ab2c-46e2-80c2-69dfdd47aefc",
		Name:         "Mind Rake",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		AlternativeCosts: []game.AlternativeCost{
			effects.Overload("{1}{B}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if !ctx.PaidAltCost("overload") {
				return effects.MindRotOnResolve(item, ctx)
			}
			ctx.Game.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
				Source: ctx.Source(), N: 2, Question: "Mind Rake — discard two cards",
			})
			return nil
		},
	})
}
