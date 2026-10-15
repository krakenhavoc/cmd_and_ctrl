package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// graveyard_offer_test.go — ADR 0187 §7 (owner decision 3, #3571): the bot
// as an offeree of Carrion Rats' "any player may exile a card from their
// graveyard". It never pays against its own Rats, pays when the Rats are
// fighting it, and declines when they are not.

// carrionOffer ranks the bot at `me` answering the Rats' offer, given the
// battlefield. The bot has one card in its graveyard.
func carrionOffer(t *testing.T, me int, bf ...protocol.CardView) (pay, decline float64) {
	t.Helper()
	const id = "55555555-5555-5555-5555-555555555555"
	fuel := cardID(90)
	gy := protocol.CardView{InstanceID: fuel, Name: "Filler", TypeLine: "Sorcery", Owner: seatID(me).String(), Controller: seatID(me).String(), KnownByYou: true}
	seats := []protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)}
	seats[me] = newSeat(me, withGraveyard(gy))
	one := 1
	v := newView(seats, withBattlefield(bf...), withTurn(5, 0, "declare_attackers"), withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_unless", Chooser: seatID(me).String(), Source: cardID(1),
		PayCost:  "exile a card from your graveyard",
		PayCards: &protocol.PayCardsView{Action: "exile_graveyard", Count: one, Options: []string{fuel}},
		AnyPlayerOffer: &protocol.AnyPlayerOfferView{
			Controller: seatID(0).String(),
			Order:      []string{seatID(0).String(), seatID(1).String(), seatID(2).String()},
		},
	}))
	in := input(me, v,
		choiceMove(t, me, id, "pay", map[string]any{"apply": true, "card_ids": []string{fuel}}),
		choiceMove(t, me, id, "decline", map[string]any{"apply": false}))
	pol := heuristic.New()
	return rankValue(t, pol, in, "pay"), rankValue(t, pol, in, "decline")
}

func TestABotNeverExilesFromItsGraveyardAgainstItsOwnRats(t *testing.T) {
	rats := creature(cardID(1), 0, "Carrion Rats", 2, 1, attacking(1))
	if pay, decline := carrionOffer(t, 0, rats); pay >= decline {
		t.Errorf("own Rats: pay %.3f, decline %.3f; the controller must decline", pay, decline)
	}
}

func TestABotBeingAttackedByTheRatsExilesFromItsGraveyard(t *testing.T) {
	rats := creature(cardID(1), 0, "Carrion Rats", 2, 1, attacking(1))
	if pay, decline := carrionOffer(t, 1, rats); pay <= decline {
		t.Errorf("attacked: pay %.3f, decline %.3f; the bot must pay", pay, decline)
	}
}

// The bot not in that combat keeps its graveyard.
func TestABotOutsideTheRatsCombatDeclines(t *testing.T) {
	rats := creature(cardID(1), 0, "Carrion Rats", 2, 1, attacking(1))
	if pay, decline := carrionOffer(t, 2, rats); pay >= decline {
		t.Errorf("not in combat: pay %.3f, decline %.3f; the bot must decline", pay, decline)
	}
}

// The Rats blocking the bot's attacker, or blocked by the bot's creature,
// is the bot's combat too; a block on someone else's creature is not.
func TestABotPaysWhenTheRatsBlockItsCreatureOrAreBlockedByIt(t *testing.T) {
	mineAttacking := creature(cardID(5), 1, "Bear", 2, 2, attacking(0))
	blockingMine := creature(cardID(1), 0, "Carrion Rats", 2, 1, blocking(cardID(5)))
	if pay, decline := carrionOffer(t, 1, mineAttacking, blockingMine); pay <= decline {
		t.Errorf("blocking my creature: pay %.3f, decline %.3f; the bot must pay", pay, decline)
	}
	if pay, decline := carrionOffer(t, 2, mineAttacking, blockingMine); pay >= decline {
		t.Errorf("blocking someone else's: pay %.3f, decline %.3f; the bot must decline", pay, decline)
	}

	attackingRats := creature(cardID(1), 0, "Carrion Rats", 2, 1, attacking(2))
	myBlocker := creature(cardID(6), 1, "Wall", 0, 4, blocking(cardID(1)))
	if pay, decline := carrionOffer(t, 1, attackingRats, myBlocker); pay <= decline {
		t.Errorf("blocked by my creature: pay %.3f, decline %.3f; the bot must pay", pay, decline)
	}
}

// The Rats aimed at a planeswalker the bot controls are fighting the bot.
func TestABotPaysWhenTheRatsAttackItsPlaneswalker(t *testing.T) {
	walker := creature(cardID(7), 1, "Walker", 0, 0)
	walker.TypeLine = "Legendary Planeswalker — Test"
	rats := creature(cardID(1), 0, "Carrion Rats", 2, 1, func(c *protocol.CardView) {
		c.AttackingTarget = walker.InstanceID
		c.AttackingTargetKind = "planeswalker"
	})
	if pay, decline := carrionOffer(t, 1, rats, walker); pay <= decline {
		t.Errorf("attacking my planeswalker: pay %.3f, decline %.3f; the bot must pay", pay, decline)
	}
	if pay, decline := carrionOffer(t, 2, rats, walker); pay >= decline {
		t.Errorf("someone else's planeswalker: pay %.3f, decline %.3f; the bot must decline", pay, decline)
	}
}
