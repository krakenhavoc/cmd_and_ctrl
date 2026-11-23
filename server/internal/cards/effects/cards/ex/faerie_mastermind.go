package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Mastermind — Creature — Faerie Rogue {1}{U}, 2/1
// (EDHREC rank 291):
//
//	"Flash
//	 Flying
//	 Whenever an opponent draws their second card each turn, you
//	 draw a card.
//	 {3}{U}: Each player draws a card."
//
// The trigger is opponent_politics.go's own named example —
// opponentDrewTheirSecondCardThisTurn was built for this card and
// Gixian Puppeteer alike. The activated ability reuses
// b05EachPlayerDraws, the shared "each player draws N" body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a984db23-40ea-428d-829f-e944267280f8",
		Name:            "Faerie Mastermind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "Faerie Mastermind — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{U}: Each player draws a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B05EachPlayerDraws(g, item, 1)
				},
			},
		},
	})
}
