package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Folio of Fancies — Artifact — Book {1}{U}:
//
//	"Players have no maximum hand size.
//	 {X}{X}, {T}: Each player draws X cards.
//	 {2}{U}, {T}: Each opponent mills cards equal to the number of cards
//	 in their hand."
//
// "Players have no maximum hand size" reaches every player (ADR 0113
// §3, #2074), folded in CR 613.11 timestamp order. {X}{X} is paid as
// twice X (the 2019-10-04 ruling) and read back with ctx.X(); each
// player draws in turn order from the active player, and a player who
// can't draw them all loses as usual. The mill counts each opponent's
// hand as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9030d508-5cfe-46ed-954b-21d970bc3b5a",
		Name:         "Folio of Fancies",
		Completeness: effects.CompletenessFull,
		HandSize:     []game.HandSizeStatic{effects.PlayersHaveNoMaxHandSize()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}{X}, {T}: Each player draws X cards.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{X}{X}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B05EachPlayerDraws(g, item, effects.NewContext(g, item).X())
				},
			},
			{
				Label:   "{2}{U}, {T}: Each opponent mills cards equal to the number of cards in their hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					// Each opponent's hand is counted as the ability resolves,
					// before anyone mills; the libraries go one at a time in
					// APNAP order (#4266).
					ctx := effects.NewContext(g, item)
					return effects.MillEachPlayerInTurn(ctx, ctx.Opponents(), func(p uuid.UUID) int { return effects.B14HandSize(g, p) }, game.KeyedTail{})
				},
			},
		},
	})
}
