package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karn's Temporal Sundering — Legendary Sorcery {4}{U}{U} (Dominaria, S82 OLD,
// #5530):
//
//	"(You may cast a legendary sorcery only if you control a legendary creature
//	 or planeswalker.)
//	 Target player takes an extra turn after this one. Return up to one target
//	 nonland permanent to its owner's hand. Exile Karn's Temporal Sundering."
//
// The cast gate is the legendary-sorcery one (CR 205.4e). Two target clauses,
// each judged on its own at resolution (CR 608.2b): a player who left costs the
// card only the extra turn, a permanent that left only the bounce (Rise of the
// Eldrazi's shape). The extra turn is the TARGET player's (CR 500.7), and the
// spell exiles itself last, so it is not left in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "5afbd367-19c7-418f-994d-a7958fb1c4ae",
		Name:               "Karn's Temporal Sundering",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.LegendarySorcery(),
		CastConditionLabel: effects.LegendarySorceryLabel,
		Purpose:            game.Purpose{ExtraTurns: 1},
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetPermanent("up to one target nonland permanent", effects.Nonland()).WithCount(0, 1),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return effects.TakeExtraTurn{Player: t.ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.ClauseTargets(1) {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					return effects.BounceToHand{Target: t.ID}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
