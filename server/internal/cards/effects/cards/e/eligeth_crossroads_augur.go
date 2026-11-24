package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eligeth, Crossroads Augur — Legendary Creature — Sphinx {4}{U}{U}, 5/6
// (Commander Legends, #5507):
//
//	"Flying
//	 If you would scry a number of cards, draw that many cards instead.
//	 Partner"
//
// A CR 614 replacement on the scry keyword action, scoped to the source's
// controller: the scry is cancelled and a body registered by key
// (game.RegisterKeywordInsteadCount, a seam beside Retriever Phoenix's
// keyword-instead that is told the action's count) draws that many cards
// instead, then runs the rest of the sentence the scry was part of ("scry 2,
// then draw a card" still draws). The count is the one the window had settled
// on, so a "scry one more" effect applied first is drawn too. Surveil is not
// scry and is left alone. Partner is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65e15354-4258-411f-a8f1-f64f4ccb873b",
		Name:            "Eligeth, Crossroads Augur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventKeywordAction},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventKeywordAction && ev.KeywordAction == game.KeywordActionScry &&
					ev.KeywordActionCount > 0 && src != nil && ev.Actor == src.Controller
			},
			KeywordInstead: game.RegisterKeywordInsteadCount("eligeth-draw-instead-of-scry", eligethDrawInsteadOfScry),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:    "Eligeth, Crossroads Augur: draw that many cards instead of scrying",
			RowLabel: "If you would scry a number of cards, draw that many cards instead.",
		}},
	})
}

// eligethDrawInsteadOfScry draws the scry's count for the scrying player,
// then lets the rest of the sentence run.
func eligethDrawInsteadOfScry(g *game.Game, actor, _ uuid.UUID, n int, done func(*game.Game) error) error {
	if err := g.DrawNForEffect(actor, n); err != nil {
		return err
	}
	return done(g)
}
