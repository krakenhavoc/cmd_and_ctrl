package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Melded Moxite — Artifact {1}{R} (Edge of Eternities, #4461):
//
//	"When this artifact enters, you may discard a card. If you do, draw
//	 two cards.
//	 {3}, Sacrifice this artifact: Create a tapped 2/2 colorless Robot
//	 artifact creature token."
//
// The entry trigger is Ragged Short Spear's (mayDiscardACardToDrawTwo):
// the discard is chosen as it resolves, up to one card, and only a card
// really discarded draws the two. The activated ability has no timing
// restriction, so the Robot can be made at instant speed; it enters
// tapped, the way Automated Assembly Line's does (TokenEntryOptions), so
// it can't block the turn it is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86e0e718-82b5-4c99-9837-1b4d8729a27e",
		Name:         "Melded Moxite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Melded Moxite — you may discard a card. If you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return mayDiscardACardToDrawTwo("Melded Moxite — discard a card to draw two?")(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice this artifact: Create a tapped 2/2 colorless Robot artifact creature token.",
			Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				robot := NewContext(g, item).TextToken(TokenCard("2/2 colorless Robot artifact"))
				_, err := g.CreateTokensForEffect(item.Controller, robot, 1, game.TokenEntryOptions{Tapped: true})
				return err
			},
		}},
	})
}
