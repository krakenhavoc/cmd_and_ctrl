package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pure Intentions — Instant — Arcane {W}:
//
//	"Whenever a spell or ability an opponent controls causes you to
//	 discard cards this turn, return those cards from your graveyard to
//	 your hand.
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, return this card from your graveyard to your hand at the
//	 beginning of the next end step."
//
// Two abilities, as printed.
//
// The spell creates a CR 603.7b delayed trigger for the rest of the
// turn: it fires on every discard an opponent's spell or ability makes
// its controller suffer (DiscardCausedByOpponentOf, the stack item's
// controller, CR 113.8) and returns the card that discard named, if it
// is still in a graveyard when the ability resolves. A card a
// replacement put onto the battlefield (Obstinate Baloth) is not in a
// graveyard, so it is left where it is. The engine emits one discard per
// card, so a two-card Mind Rot fires it twice and returns both, as
// "those cards" does.
//
// The second is the family's "this card" trigger (ADR 0153 §4), watching
// from wherever the card landed (WhereverDiscarded, #3859), scheduling a delayed trigger for the next end step
// that returns the card, pinned as an object (CR 400.7). It is not
// Pure Intentions' own first ability: that one is only the spell's
// effect, so a Pure Intentions that is discarded returns on its own.
//
// No simplification: a card that madness or Rest in Peace sent to exile
// still triggers it, and the delayed return then finds nothing in a
// graveyard to return, as "from your graveyard" prints.
func init() {
	Register(Spec{
		OracleID:          "77ef82e6-16e3-44fa-9542-252060fa6c96",
		Name:              "Pure Intentions",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DelayedOnEvent{
				Label:     "Pure Intentions — return the discarded card from your graveyard to your hand",
				On:        []game.EventKind{game.EventDiscardCard},
				Condition: opponentCausedYourDiscardCondition,
				Body:      returnTheDiscardedCardBody,
				Repeats:   true,
			}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			WhereverDiscarded(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Pure Intentions — discarded: return it to your hand at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return scheduleReturnOfThisCard(g, item,
						"Pure Intentions — return it from your graveyard to your hand", returnDiscardedToHandBody)
				})),
		},
	})
}
