package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gustha Ebbasdotter — Legendary Creature — Human Wizard {3}{U}{U}, 3/3
// (Mystery Booster Commander, #5948):
//
//	"Protection from red
//	 Whenever you cast your first spell during each opponent's turn, draw
//	 three cards, then put two cards from your hand on top of your library
//	 in any order."
//
// The condition is Alela's (b22FirstSpellOnAnOpponentsTurn): a cast of yours
// during a turn that is not yours, and the first one of that turn. The
// effect is Brainstorm's, through the keyed draw-then put-back Brainstone
// uses, with Gustha's own prompt label.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0a8c04e-3ce7-4945-a6af-7e40a990c82c",
		Name:            "Gustha Ebbasdotter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("R")},
		Purpose:         game.Purpose{Draws: 3, Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b22FirstSpellOnAnOpponentsTurn(ev, source, g)
			}, "Gustha Ebbasdotter — draw three cards, then put two cards from your hand on top of your library in any order",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
						Ref: gapB05GusthaPutBack, Player: item.Controller, Source: item.SourceCardID, N: 2,
					})
				}),
		},
	})
}

// gapB05GusthaPutBack is "then put two cards from your hand on top of your
// library in any order" with Gustha's prompt label. An on-disk key.
var gapB05GusthaPutBack = game.RegisterDrawThen("gap-b05-gustha-put-back", func(g *game.Game, d game.DrawThen) error {
	ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
	return PutFromHandOnTopInAnyOrder{
		Player: d.Player,
		N:      d.N,
		Label:  "Gustha Ebbasdotter — put two cards from your hand on top of your library",
	}.Apply(ctx)
})
