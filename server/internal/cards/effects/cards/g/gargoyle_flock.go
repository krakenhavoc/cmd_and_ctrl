package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gargoyle Flock — Creature — Tyranid Gargoyle {2}{G}{U}, 2/2 (Warhammer
// 40,000 Commander, S82 GAP batch 193, #5948):
//
//	"Flying
//	 Skyswarm — At the beginning of your end step, if a creature entered the
//	 battlefield under your control this turn, create a 1/1 blue Tyranid
//	 Gargoyle creature token with flying."
//
// The condition is an intervening "if" (CR 603.4) on the turn's creature
// entry tally (Wary Farmer's reading), the Flock itself counting: checked as
// the end step begins and again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6d24fc3e-eaa9-4ac8-b271-7dc83b812acf",
		Name:            "Gargoyle Flock",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "a creature entered the battlefield under your control this turn",
				Test: func(ctx *effects.Context) bool {
					return effects.GapB193YouControlACreatureEntered(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && effects.GapB193YouControlACreatureEntered(g, source.Controller)
			}, "Gargoyle Flock — create a 1/1 blue Tyranid Gargoyle creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Tyranid Gargoyle with flying"), N: 1}))),
		},
	})
}
