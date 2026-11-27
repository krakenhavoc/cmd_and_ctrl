package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thraben Sentry // Thraben Militia — Creature — Human Soldier {3}{W},
// 2/2 (Innistrad, S82 OLD batch 450, #5572):
//
//	"Vigilance
//	 Whenever another creature you control dies, you may transform this
//	 creature."
//	// Thraben Militia (back): 5/4 Trample
//
// Gatstaf Shepherd's two registrations (the back face under "#1"), the
// trigger of Blood Artist's dies shape over "another creature you control"
// with Vildin-Pack Alpha's optional transform. Transforming a Sentry that
// has itself just died does nothing.
//
// No simplification.
func init() {
	const oracle = "9636073e-e361-4d2a-92f6-5fe049e5eb23"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Thraben Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureYouControlDied(ev, source, g)
			}, "Thraben Sentry — you may transform this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Thraben Sentry — transform it?",
						Then:     thrabenSentryTransform,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Thraben Militia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}

var thrabenSentryTransform = effects.MayYes("confirm/s82-old-b450-thraben-sentry-transform", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.TransformThis{}.Apply(ctx)
})
