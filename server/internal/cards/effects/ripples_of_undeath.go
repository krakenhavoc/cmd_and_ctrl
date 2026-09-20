package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ripples of Undeath — Enchantment {1}{B}:
//
//	"At the beginning of your first main phase, mill three cards.
//	 Then you may pay {1} and 3 life. If you do, put a card from
//	 among those cards into your hand."
//
// A two-mana engine that fills a graveyard every turn and, for a
// price, turns the best of the three into a draw. In a deck that
// wants cards in the graveyard the mill is the upside and the
// payment is optional, which is what makes it playable at two mana.
//
// # Three mechanisms, in the order the card prints them
//
//   - "Your FIRST main phase" is the precombat main (CR 505.1a), so
//     the trigger is AtYourPrecombatMain and it fires once a turn on
//     its controller's turn only. A postcombat main phase is not a
//     first main phase and does not fire it again.
//   - The mill has a CONTINUATION, because "those cards" is the set
//     that actually reached the graveyard. A card a replacement sent
//     elsewhere — a commander that took CR 903.9's offer, an "exile
//     it instead" effect — was not milled and is not a candidate, and
//     the mill can pause on that prompt, so the list is not knowable
//     on the line after the mill (MillToZone.Then, #893).
//   - "You may pay {1} and 3 life" is an optional cost paid DURING
//     RESOLUTION, which is MayPay.
//
// # The cost is split, and why that is faithful
//
// The engine's may-pay prompt carries a parsed MANA cost and nothing
// else — a ParsedCost has no life component — so the {1} rides the
// prompt and the 3 life is paid inside OnPay, before the card is
// taken. Two things keep that honest:
//
//   - CR 119.4 says a player may pay 3 life only at a life total of
//     at least 3, so the offer is not made below it. A pending choice
//     holds the table until it is answered, so nothing can change the
//     life total between the check and the payment.
//   - The life is paid through the COST path (PayLifeForEffect), so
//     it is a real payment that a life-loss watcher sees, and it is
//     paid BEFORE the card moves — the order the card prints.
//
// The question names both halves so the player is never asked about
// half a cost.
//
// An empty library mills nothing and the offer is still made: "you
// may pay" is not conditional on the mill having found cards, and
// paying into an empty set is a legal (if terrible) choice. Nothing
// is then put into hand, because there is nothing to put.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2acf8e34-9215-4a72-a24f-09d3bdd0083a",
		Name:         "Ripples of Undeath",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain(ripplesOfUndeathLabel, ripplesOfUndeathMillThenOffer),
		},
	})
}

// ripplesOfUndeathLabel is the stack label of the one printed
// ability.
const ripplesOfUndeathLabel = "Ripples of Undeath — mill three, then you may pay {1} and 3 life"

// ripplesOfUndeathLife is the life half of the optional cost.
const ripplesOfUndeathLife = 3

// ripplesOfUndeathMillThenOffer mills three and hands the cards that
// actually arrived to the optional-cost offer.
func ripplesOfUndeathMillThenOffer(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		Player: item.Controller,
		N:      3,
		To:     game.ZoneGraveyard,
		Then:   ripplesOfUndeathOffer,
	}.Apply(NewContext(g, item))
}

// ripplesOfUndeathOffer asks the controller whether to pay. The life
// half is checked here rather than in OnPay: CR 119.4 makes a payment
// the player cannot afford no payment at all, and an offer nobody can
// take is not an offer.
func ripplesOfUndeathOffer(ctx *Context, milled []uuid.UUID) error {
	controller := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p == nil || p.Life < ripplesOfUndeathLife {
		return nil
	}
	cards := append([]uuid.UUID(nil), milled...)
	return MayPay{
		Chooser:  controller,
		Cost:     "{1}",
		Question: "Ripples of Undeath — pay {1} and 3 life to put one of the milled cards into your hand?",
		OnPay: func(ctx *Context) error {
			return ripplesOfUndeathTakeOne(ctx, cards)
		},
	}.Apply(ctx)
}

// ripplesOfUndeathTakeOne pays the life half and asks which of the
// milled cards to keep. The life goes first: the card prints the
// payment before the card moves, and a Blood Artist style payoff
// watching the loss has to see it in that order.
func ripplesOfUndeathTakeOne(ctx *Context, milled []uuid.UUID) error {
	player := ctx.Controller()
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), player, ripplesOfUndeathLife); err != nil {
		return err
	}
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, id := range milled {
		if p.Graveyard.Contains(id) {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	resolving := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Ripples of Undeath — put a card from among the milled cards into your hand",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		// Re-checked on submit: a card can leave the graveyard
		// between the question and the answer.
		Zone: game.ZoneGraveyard,
		Then: ReturnPickedToHand(resolving),
	})
	return nil
}
