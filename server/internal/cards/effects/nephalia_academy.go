package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nephalia Academy — Land:
//
//	"If a spell or ability an opponent controls causes you to discard a
//	 card, you may reveal that card and put it on top of your library
//	 instead of putting it anywhere else.
//	 {T}: Add {C}."
//
// ADR 0153 §1, the library half of the family: the same cause test as
// the Baloth (the stack item's controller, CR 113.8), applied to ANY
// card the Academy's controller discards, so it is the permanent that
// carries the replacement rather than the discarded card.
//
// "You may" is the replacement's Optional prompt, asked per card, as
// Library of Leng's is. "Reveal that card" runs when the player says
// yes, before the move: the whole table sees the card in the hand it is
// leaving, and the reveal is the existing reveal event. Declining
// leaves the discard to land in the graveyard. The card was discarded
// either way, so a "whenever you discard" payoff sees it, with the zone
// it landed in.
//
// "Instead of putting it anywhere else" is DiscardBecomes' guard that
// the event still be headed for the graveyard (CR 614.5): once madness
// or Rest in Peace has moved the card, the Academy no longer applies.
//
// The card declares `library` for every card in its controller's hand,
// so the bot knows that a discard an opponent forces is not a loss
// while the Academy is out (ADR 0153 §5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                 "3b7e7a11-bf59-413d-8796-640d17c2c1c6",
		Name:                     "Nephalia Academy",
		Completeness:             CompletenessFull,
		DiscardByOpponentForHand: game.DiscardByOpponentLibrary,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Replacements: []game.ReplacementEffect{
			DiscardBecomes{
				Dst:        game.ZoneLibrary,
				ByOpponent: true,
				Optional:   true,
				OnApply: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					g.RevealForEffect(game.RevealSpec{
						Player: ev.DiscardPlayer,
						Source: src.InstanceID,
						Reason: "Nephalia Academy — reveal the discarded card",
						Cards:  []uuid.UUID{ev.CardID},
					})
					return nil
				},
				Label:    "Nephalia Academy: reveal it and put it on top of your library instead",
				Question: "Reveal the discarded card and put it on top of your library instead?",
			}.Build(),
		},
	})
}
