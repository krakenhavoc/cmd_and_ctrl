package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mangara's Blessing — Instant {2}{W}:
//
//	"You gain 5 life.
//	 When a spell or ability an opponent controls causes you to discard
//	 this card, you gain 2 life, and you return this card from your
//	 graveyard to your hand at the beginning of the next end step."
//
// The family's "this card" trigger (ADR 0153 §4), watching from the
// GRAVEYARD on a discard an opponent's spell or ability caused (the stack
// item's controller, CR 113.8). It gains 2 life at once and schedules a
// delayed trigger for the next end step that returns the card, pinned as
// an object (CR 400.7): a card that has left the graveyard in the
// meantime is not dragged back. "You" is the card's owner while it sits
// there (CR 108.4).
//
// Caveat, declared: if madness or Rest in Peace sends the discarded card
// to exile instead of the graveyard, its ability does not trigger from
// exile (owner decision 6).
func init() {
	Register(Spec{
		OracleID:          "5d9431de-7cf0-43d4-9b86-e0ef4a7756f9",
		Name:              "Mangara's Blessing",
		Completeness:      CompletenessCaveats,
		Caveats:           []string{discardTriggerFromExileCaveat},
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Purpose:           game.Purpose{LifeGain: 5},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainLife{Amount: 5}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Mangara's Blessing — discarded: gain 2 life, return it to your hand at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					if err := (GainLife{Amount: 2}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					return scheduleReturnOfThisCard(g, item,
						"Mangara's Blessing — return it from your graveyard to your hand", returnDiscardedToHandBody)
				})),
		},
	})
}
