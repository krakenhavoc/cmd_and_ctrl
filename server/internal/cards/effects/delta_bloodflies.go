package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Delta Bloodflies — Creature — Insect {1}{B}, 1/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Flying
//	 Whenever this creature attacks, if you control a creature with a
//	 counter on it, each opponent loses 1 life."
//
// The attack trigger's intervening "if" (CR 603.4) is checked as the
// creature attacks and again as the ability resolves: a counter of any
// kind, on any creature you control, the Bloodflies included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a64cf8e-ae7f-4697-93fc-9c946eb1088b",
		Name:            "Delta Bloodflies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control a creature with a counter on it",
				Test: func(ctx *Context) bool { return tdmControlsCreatureWithACounter(ctx.Game, ctx.Controller()) },
			}, On(game.EventAttack, AllOf(ThisAttacked, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return tdmControlsCreatureWithACounter(g, source.Controller)
			}), "Delta Bloodflies — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1}))),
		},
	})
}
