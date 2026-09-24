package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Library — Enchantment {1}{G}:
//
//	"At the beginning of your draw step, you may draw two additional
//	 cards. If you do, choose two cards in your hand drawn this turn.
//	 For each of those cards, pay 4 life or put the card on top of
//	 your library."
//
// The card that named the gap. It is three decisions deep and each one
// needs the answer to the one before it, which is a shape the choice
// queue had no composition for until #74's chained-choice work:
//
//  1. "you may draw two additional cards" — the trigger's own CR 603.5
//     optional prompt, which the harvester has offered since S19.
//  2. "choose two cards in your hand drawn this turn" — a card-set pick
//     over a candidate list that does not exist until step 1 has been
//     answered yes and the two extra cards are in hand.
//  3. "for each of those cards, pay 4 life or put the card on top of
//     your library" — one two-way prompt per card CHOSEN IN STEP 2. The
//     prompts are not merely queued after the pick; they are queued out
//     of its answer, and there is no way to know how many there will be
//     or what they will say until it arrives.
//
// Step 3 is a chain in its own right: the second card's prompt is
// queued by the first card's answer rather than both at once, so the
// player decides the second payment already knowing what the first one
// cost them. At 8 life, "pay 4" twice is lethal and "pay 4" once is
// not, and a player who saw both questions side by side would be
// answering the second one before the first had happened.
//
// # Free, and that is the card
//
// Nothing here is free in the mana sense and everything is free in the
// life sense: Sylvan Library is a two-mana enchantment that draws three
// cards a turn for 8 life, and in Commander that is a real rate. The
// engine does not need to know any of that. What it needs to get right
// is that the life is paid per card and the put-back goes on TOP (the
// card comes straight back next turn, which is why "put back" is a soft
// cost rather than a discard).
//
// # Declared sandbox simplification: PAYING IS OFFERED ONLY WHEN LEGAL
//
// CR 119.4 lets a player pay life only down to 0, so "pay 4" is offered
// only when the payer has at least 4 life; below that the card is put
// back without a prompt, which is the only legal outcome anyway. A
// player at exactly 4 CAN still pay and lose to the state-based action,
// because that is a legal (and occasionally correct) play and the
// engine has no business second-guessing it.
func init() {
	Register(Spec{
		OracleID:     "92eed395-62ca-4293-882b-8565c40daab5",
		Name:         "Sylvan Library",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// "You MAY draw two additional cards" — the first link of
			// the chain, and the one the engine already had.
			Optional(On(game.EventBeginDrawStep, ByYou, "Sylvan Library — draw two additional cards", func(g *game.Game, item *game.StackItem) error {
				controller := item.Controller
				if err := g.DrawNForEffect(controller, 2); err != nil {
					return err
				}
				// "choose two cards in your hand drawn this
				// turn" — every card drawn this turn, not just
				// the two this ability drew. The turn-based
				// draw is eligible and is frequently the right
				// card to put back, which is why the engine
				// tracks the draws rather than this file
				// remembering its own two.
				drawn := g.CardsDrawnThisTurnFor(controller)
				if len(drawn) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  controller,
					Source:   item.SourceCardID,
					Question: "Sylvan Library — choose two cards drawn this turn",
					Cards:    drawn,
					// "Choose two" — and no fewer, unless the
					// player has fewer than two left to
					// choose from (they discarded one, or an
					// opponent's effect took it).
					Min:  min(2, len(drawn)),
					Max:  min(2, len(drawn)),
					Zone: game.ZoneHand,
					Then: func(g *game.Game, picked []uuid.UUID) error {
						return sylvanLibrarySettle(g, item.SourceCardID, controller, picked)
					},
				})
				return nil
			}), "Sylvan Library — draw two additional cards?"),
		},
	})
}

// sylvanLibrarySettle asks about the first of the remaining chosen
// cards, and queues itself again — via the answer's continuation — for
// the rest. One prompt at a time, in the order the player chose them.
//
// Recursion through the queue rather than a loop is the whole point:
// the second question cannot be asked until the first has been
// answered, because the answer to the first is what the player is
// pricing the second against.
//
// Caller must hold g.mu.
func sylvanLibrarySettle(g *game.Game, source, controller uuid.UUID, remaining []uuid.UUID) error {
	if len(remaining) == 0 {
		return nil
	}
	card := remaining[0]
	rest := remaining[1:]
	putBack := func(g *game.Game) error {
		// Top, not bottom: the card is back on the library and will be
		// drawn again next turn. That is what makes the put-back leg a
		// tempo cost rather than a loss.
		//
		// #783: the NEXT question hangs off the tuck's continuation
		// rather than the next line, for the same reason the first one
		// hangs off the answer to the one before it — a library is a
		// CR 903.9 destination, so putting a drawn commander back stops
		// to ask its owner about the command zone, and the rest of the
		// card must not run while that is open.
		return g.TuckToLibraryThenForEffect(card, game.TuckOptions{}, func(g *game.Game, _ bool) error {
			return sylvanLibrarySettle(g, source, controller, rest)
		})
	}
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Life < sylvanLibraryLifeCost {
		// CR 119.4: a player cannot pay life they do not have, so
		// there is no question to ask. Put it back and move on rather
		// than offering a branch the engine would have to refuse —
		// an unanswerable prompt is how a table wedges (#544).
		return putBack(g)
	}
	name := "that card"
	if c, ok := g.LookupCardForEffect(card); ok && c.Name != "" {
		name = c.Name
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      controller,
		Source:       source,
		Question:     fmt.Sprintf("Sylvan Library — %s: pay %d life to keep it?", name, sylvanLibraryLifeCost),
		AcceptLabel:  fmt.Sprintf("Pay %d life", sylvanLibraryLifeCost),
		DeclineLabel: "Put it on top",
		// Declared so the move list prices it. Without it a bot at 4
		// life answers "pay" and dies — the same hole #547 closed for
		// activated abilities, reached through a prompt instead.
		LifeCost: sylvanLibraryLifeCost,
		OnAccept: func(g *game.Game) error {
			// Life can move between the question and the answer (an
			// opponent's burn spell in response to nothing in
			// particular), so the payment is re-checked rather than
			// trusted from prompt time. A payment that can no longer
			// be made degrades to the put-back, which is the same
			// shape ResolveEntryPayLife uses for a shockland.
			if p := g.PlayerByIDForEffect(controller); p == nil || p.Life < sylvanLibraryLifeCost {
				return putBack(g)
			}
			// A cost (CR 118.3), so the cost path: the CR 614 window
			// runs on it (CR 119.4) but settles in one step, which
			// keeps the per-card chain moving rather than stranding
			// the rest of it behind a CR 616 prompt (#793).
			if err := g.PayLifeForEffect(source, controller, sylvanLibraryLifeCost); err != nil {
				return err
			}
			// Keeping the card is the whole of the "pay" branch — it
			// is already in hand. The chain continues to the next
			// chosen card.
			return sylvanLibrarySettle(g, source, controller, rest)
		},
		OnDecline: putBack,
	})
	return nil
}

// sylvanLibraryLifeCost is the printed 4.
const sylvanLibraryLifeCost = 4
