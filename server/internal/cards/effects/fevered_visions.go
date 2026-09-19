package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fevered Visions — Enchantment for {1}{U}{R} (EDHREC rank 4439):
//
//	"At the beginning of each player's end step, that player draws a
//	 card. If the player is your opponent and has four or more cards
//	 in hand, this enchantment deals 2 damage to that player."
//
// A symmetric Howling Mine with an asymmetric bill: everyone draws,
// but only opponents who hoard take the two. The card is a clock that
// punishes the exact behaviour drawing extra cards encourages, and in
// a four-player game it ticks three times a round. Roadmap batch 42
// (#449), "no new machinery".
//
// "EACH player's end step" is the whole table's, so the trigger
// watches every end step rather than only its controller's — and the
// player in question is the one whose step it is, read off the event's
// actor when the trigger is built. Writing it as the controller would
// make the enchantment draw its owner a card on every opponent's turn
// and never draw the opponents anything, which is not this card.
//
// The DRAW is unconditional and includes you; the DAMAGE is
// conditional on two things checked when the trigger RESOLVES, after
// the draw has happened: the player is an opponent, and their hand is
// four or more. So a player at exactly three cards who draws to four
// takes the two — the draw is part of the same ability and happens
// first, as printed.
//
// DAMAGE from the enchantment, not life loss: it is preventable, and
// it is the enchantment that dealt it, which matters for anything
// that cares who the source was.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70763549-4b4e-4cb8-8c02-0639ba18bb1a",
		Name:         "Fevered Visions",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				who := ev.Actor
				return game.NewTriggeredItem(source, "Fevered Visions — that player draws, and takes 2 with four or more in hand",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (DrawCards{Player: who, N: 1}).Apply(ctx); err != nil {
							return err
						}
						if who == item.Controller {
							return nil
						}
						p := g.PlayerByIDForEffect(who)
						if p == nil || len(p.Hand.Cards) < 4 {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: who, Amount: 2}.Apply(ctx)
					})
			},
		}},
	})
}
