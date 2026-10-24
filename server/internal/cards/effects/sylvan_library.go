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
			OptionalDraw(On(game.EventBeginDrawStep, ByYou, "Sylvan Library — draw two additional cards", Do(
				// The choice of cards waits for a prompt the draws ask, such as a
				// dredge offer: it is made over what was drawn (CR 608.2c, #4155).
				Instruction(func(ctx *Context) error {
					return ctx.Game.DrawNForEffect(ctx.Controller(), 2)
				}),
				Instruction(func(ctx *Context) error {
					g, item := ctx.Game, ctx.Item
					controller := item.Controller
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
						Min: min(2, len(drawn)),
						Max: min(2, len(drawn)),
						// #4301: keyed, so a table waiting on the pick is a
						// restore point. The cards must still be in the
						// hand when the answer arrives (LiveZone).
						ThenKey:  sylvanLibraryChosen,
						LiveZone: game.ZoneHand,
					})
					return nil
				}),
			)), "Sylvan Library — draw two additional cards?", 2),
		},
	})
	sylvanLibraryPay = RegisterThenSteps("then-steps/sylvan-library-pay", sylvanLibraryPaySteps()...)
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
// The loop's progress is data on the prompt (#4298): the cards still to
// settle, the one asked about first, are its carried IDs, so a table
// waiting on a payment is a restore point.
//
// Caller must hold g.mu.
func sylvanLibrarySettle(ctx *Context, remaining []uuid.UUID) error {
	g := ctx.Game
	if len(remaining) == 0 || ctx.Item == nil {
		return nil
	}
	controller := ctx.Controller()
	card := remaining[0]
	rest := remaining[1:]
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Life < sylvanLibraryLifeCost {
		// CR 119.4: a player cannot pay life they do not have, so
		// there is no question to ask. Put it back and move on rather
		// than offering a branch the engine would have to refuse —
		// an unanswerable prompt is how a table wedges (#544).
		return sylvanLibraryPutBack(ctx, card, rest)
	}
	name := "that card"
	if c, ok := g.LookupCardForEffect(card); ok && c.Name != "" {
		name = c.Name
	}
	return MayChoice{
		Player:   controller,
		Question: fmt.Sprintf("Sylvan Library — %s: pay %d life to keep it?", name, sylvanLibraryLifeCost),
		YesLabel: fmt.Sprintf("Pay %d life", sylvanLibraryLifeCost),
		NoLabel:  "Put it on top",
		// Declared so the move list prices it. Without it a bot at 4
		// life answers "pay" and dies — the same hole #547 closed for
		// activated abilities, reached through a prompt instead.
		LifeCost: sylvanLibraryLifeCost,
		Then:     sylvanLibraryAnswer,
		Carry:    remaining,
	}.Apply(ctx)
}

// sylvanLibraryPutBack puts the card on top of the library and asks about
// the rest.
func sylvanLibraryPutBack(ctx *Context, card uuid.UUID, rest []uuid.UUID) error {
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
	item := ctx.Item
	return ctx.Game.TuckToLibraryThenForEffect(card, game.TuckOptions{}, func(g *game.Game, _ bool) error {
		return sylvanLibrarySettle(NewContext(g, item), rest)
	})
}

// sylvanLibraryAnswer is the answer to one card's "pay 4 life or put it on
// top". Registered in init, because the payment's steps ask the next card
// through sylvanLibrarySettle, which names it.
var sylvanLibraryAnswer game.ConfirmThen

func init() {
	sylvanLibraryAnswer = MayAnswer("confirm/sylvan-library-pay-or-top", func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		card, rest := r.Carry[0], r.Carry[1:]
		if !r.Accepted {
			return sylvanLibraryPutBack(ctx, card, rest)
		}
		// Life can move between the question and the answer (an
		// opponent's burn spell in response to nothing in
		// particular), so the payment is re-checked rather than
		// trusted from prompt time. A payment that can no longer
		// be made degrades to the put-back, which is the same
		// shape ResolveEntryPayLife uses for a shockland.
		if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p == nil || p.Life < sylvanLibraryLifeCost {
			return sylvanLibraryPutBack(ctx, card, rest)
		}
		// Keeping the card is the whole of the "pay" branch — it
		// is already in hand. The chain continues to the next
		// chosen card, after the payment (continuation steps,
		// #4162, CR 608.2c).
		ctx.RecordStep(sylvanLibraryRecord, rest)
		return sylvanLibraryPay.Run(ctx)
	})
}

// sylvanLibraryChosen is the answer to "choose two cards in your hand drawn
// this turn" (#4301): the prompt's chooser is the controller and its source
// the card, so the chain runs against a stand-in for the long-gone trigger,
// as the payments after it do. An on-disk key.
var sylvanLibraryChosen = game.RegisterChooseCardsThen("choose-cards/sylvan-library-chosen", func(g *game.Game, r game.CardsChosen) error {
	return sylvanLibrarySettle(keyedContext(g, r.Chooser, r.Source), r.Picked)
})

// sylvanLibraryRecord is the chosen cards still to settle.
const sylvanLibraryRecord = "sylvan-library/rest"

// sylvanLibraryPay is the payment and then the next card's question. It
// is registered in init, because the next question's answer runs it
// again.
var sylvanLibraryPay ThenSteps

func sylvanLibraryPaySteps() []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			// A cost (CR 118.3), so the cost path: the CR 614 window
			// runs on it (CR 119.4) but settles in one step (#793). As
			// a step, the next question would wait for it even if it
			// asked.
			return ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), sylvanLibraryLifeCost)
		}),
		Instruction(func(ctx *Context) error {
			return sylvanLibrarySettle(ctx, ctx.StepRecord(sylvanLibraryRecord).IDs)
		}),
	}
}

// sylvanLibraryLifeCost is the printed 4.
const sylvanLibraryLifeCost = 4
