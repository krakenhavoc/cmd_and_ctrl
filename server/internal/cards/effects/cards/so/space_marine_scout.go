package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Space Marine Scout — Creature — Astartes Scout {2}{W}, 2/1 (Warhammer 40,000
// Commander, S82 GAP batch 196, #5948):
//
//	"First strike, vigilance
//	 Concealed Position — When this creature enters, if an opponent controls more
//	 lands than you, you may search your library for a Plains card, put it onto
//	 the battlefield tapped, then shuffle."
//
// Dora Milaje Elite's intervening "if" (CR 603.4) on the land counts: checked as
// the Scout enters and again as the trigger resolves, so an equalised count
// makes it do nothing. The search is a real "you may" with the shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b7fb630-79eb-4337-a698-626e585a495f",
		Name:            "Space Marine Scout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "an opponent controls more lands than you",
				Test: func(ctx *effects.Context) bool {
					return effects.B05OpponentControlsMoreLands(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.B05OpponentControlsMoreLands(g, source.Controller)
			}, "Space Marine Scout — you may search your library for a Plains card, put it onto the battlefield tapped, then shuffle",
				effects.GapB196ConcealedPosition)),
		},
	})
}
