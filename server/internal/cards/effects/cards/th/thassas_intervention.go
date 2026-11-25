package cardsth

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thassa's Intervention — Instant {X}{U}{U} (Theros Beyond Death, #5514):
//
//	"Choose one —
//	 • Look at the top X cards of your library. Put up to two of them into
//	   your hand and the rest on the bottom of your library in a random order.
//	 • Counter target spell unless its controller pays twice {X}."
//
// A choose-one with X announced at the cast. The first mode is a look (only
// the looker sees the cards, and nothing is revealed): up to two go to the
// hand, the rest to the bottom in a random order. The second is Condescend's
// pay-unless over the spell's one target, at {2X} (twice X); with the target
// illegal the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7c89d03-606e-4284-90dd-64b1fbebfb4d",
		Name:         "Thassa's Intervention",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(
				effects.Mode("Look at the top X cards of your library. Put up to two of them into your hand and the rest on the bottom of your library in a random order."),
				game.Purpose{Answers: game.AnswerValue, Draws: 2}),
			effects.ModeWithPurpose(
				effects.Mode("Counter target spell unless its controller pays twice {X}.", effects.TargetSpell("target spell")),
				game.Purpose{Answers: game.AnswerRemove}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			switch {
			case ctx.HasMode(0):
				return effects.TakeFromLibraryToHand{
					Player:   ctx.Controller(),
					Cards:    ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), x),
					Max:      2,
					Optional: true,
					Label:    "Thassa's Intervention — put up to two of them into your hand",
					Key:      effects.TakeRestBottomRandom,
				}.Apply(ctx)
			case ctx.HasMode(1):
				for _, t := range ctx.ModeTargets(0) {
					if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
						cost := "{" + strconv.Itoa(2*x) + "}"
						return effects.CounterUnlessPaid{
							StackID:  t.ID,
							Cost:     cost,
							Question: "Thassa's Intervention — pay " + cost + " or your spell is countered",
						}.Apply(ctx)
					}
				}
			}
			return nil
		},
	})
}
