package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fa'adiyah Seer — Creature — Human Shaman {1}{G}, 1/1 (Dominaria
// Remastered):
//
//	"{T}: Draw a card and reveal it. If it isn't a land card, discard it."
//
// A draw with a registered continuation (DrawThenDiscard's seam,
// DrawNThenForEffect): the card drawn is the one the player's turn tally
// grew by, so a draw that was replaced or came from an empty library
// reveals and discards nothing. The card is revealed, and if it is not a
// land card (a token is not a card, CR 108.2) it is discarded; a land stays
// in hand. The discard is that card and no other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ba2f559-f9bd-48ef-9fef-6dffe768f4ed",
		Name:         "Fa'adiyah Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card and reveal it. If it isn't a land card, discard it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				return g.DrawNThenForEffect(player, 1, game.DrawThen{
					Ref:    faadiyahSeerDrew,
					Player: player,
					Source: item.SourceCardID,
					N:      len(g.DrawnThisTurn[player]),
				})
			},
		}},
	})
}

// faadiyahSeerDrew is the rest of the sentence, after the draw. N carries
// the length of the player's draw tally before it. An on-disk key: never
// renamed, never reused.
var faadiyahSeerDrew = game.RegisterDrawThen("faadiyah-seer/drew", func(g *game.Game, d game.DrawThen) error {
	drawn := g.DrawnThisTurn[d.Player]
	if len(drawn) <= d.N {
		return nil
	}
	id := drawn[len(drawn)-1]
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	g.RevealForEffect(game.RevealSpec{
		Player: d.Player,
		Source: d.Source,
		Reason: "Fa'adiyah Seer — the drawn card",
		Cards:  []uuid.UUID{id},
	})
	if card.IsLand() {
		return nil
	}
	return g.DiscardCardsKeyedForEffect(d.Player, []uuid.UUID{id}, d.Source, game.KeyedTail{})
})
