package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghostly Pilferer — Creature — Spirit Rogue {1}{U}, 2/1 (Outlaws of Thunder
// Junction Commander):
//
//	"Whenever this creature becomes untapped, you may pay {2}. If you do,
//	 draw a card.
//	 Whenever an opponent casts a spell from anywhere other than their
//	 hand, draw a card.
//	 Discard a card: This creature can't be blocked this turn."
//
// Three separate abilities. The untap trigger is Key to the City's
// (EventUntapCard on itself, a MayPay of {2}). The cast trigger reads the
// zone the spell left (cmdr2S40OpponentCastNotFromHand): a commander from
// the command zone, a foretold or plotted card from exile, a flashback
// card all draw, a hand cast and a copy do not. The activated ability
// discards a card as its cost and makes this creature unblockable until
// end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe168bf1-f626-460c-890b-1ce758af01fc",
		Name:         "Ghostly Pilferer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return RestrictUntilEOT{
					Target:       item.SourceCardID,
					Restrictions: game.CantBeBlocked,
					Label:        "Ghostly Pilferer — can't be blocked",
				}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self, "Ghostly Pilferer — you may pay {2} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Ghostly Pilferer — pay {2} to draw a card?",
						Then:     mayPayDrawACard,
					}.Apply(NewContext(g, item))
				}),
			On(game.EventCast, cmdr2S40OpponentCastNotFromHand,
				"Ghostly Pilferer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
