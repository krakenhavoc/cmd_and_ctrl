package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// any_player_offer_test.go — ADR 0150 §6: the bot as an offeree of an
// "unless any player pays" offer. It never pays against its own
// effect, an opponent's it pays as any tax (no free-riding, owner
// decision 7), and Icy Prison's offer, where paying helps the
// controller, is the other way round.

func anyPlayerOfferValues(t *testing.T, me int, controller int, helps bool) (pay, decline float64) {
	t.Helper()
	pol := heuristic.New()
	const id = "33333333-3333-3333-3333-333333333333"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)}, withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_unless", Chooser: seatID(me).String(), PayCost: "{2}",
		AnyPlayerOffer: &protocol.AnyPlayerOfferView{
			Controller:      seatID(controller).String(),
			Order:           []string{seatID(0).String(), seatID(1).String(), seatID(2).String()},
			HelpsController: helps,
		},
	}))
	in := input(me, v,
		choiceMove(t, me, id, "pay", map[string]any{"apply": true}),
		choiceMove(t, me, id, "decline", map[string]any{"apply": false}))
	return rankValue(t, pol, in, "pay"), rankValue(t, pol, in, "decline")
}

func TestABotNeverPaysAnOfferAgainstItsOwnEffect(t *testing.T) {
	pay, decline := anyPlayerOfferValues(t, 0, 0, false)
	if pay >= decline {
		t.Errorf("own offer: pay %.3f, decline %.3f; the controller must decline", pay, decline)
	}
}

// An opponent pays even though a later opponent could still pay: no
// free-riding.
func TestABotOpponentPaysAnAnyPlayerOfferWithoutFreeRiding(t *testing.T) {
	pay, decline := anyPlayerOfferValues(t, 1, 0, false)
	if pay <= decline {
		t.Errorf("opponent's offer: pay %.3f, decline %.3f; the bot must pay", pay, decline)
	}
}

// lifeOfferValues is anyPlayerOfferValues for an opponent's offer paid
// in life or damage (ADR 0150 §6), with the bot at `life`.
func lifeOfferValues(t *testing.T, life, n int, damage bool) (pay, decline float64) {
	t.Helper()
	pol := heuristic.New()
	const id = "44444444-4444-4444-4444-444444444444"
	ch := protocol.PendingChoiceView{
		ID: id, Kind: "pay_unless", Chooser: seatID(1).String(),
		AnyPlayerOffer: &protocol.AnyPlayerOfferView{
			Controller: seatID(0).String(),
			Order:      []string{seatID(0).String(), seatID(1).String(), seatID(2).String()},
		},
	}
	if damage {
		ch.PayDamage, ch.PayCost = &n, "5 damage"
	} else {
		ch.PayLife, ch.PayCost = &n, "5 life"
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1, withLife(life)), newSeat(2)}, withChoice(ch))
	in := input(1, v,
		choiceMove(t, 1, id, "pay", map[string]any{"apply": true}),
		choiceMove(t, 1, id, "decline", map[string]any{"apply": false}))
	return rankValue(t, pol, in, "pay"), rankValue(t, pol, in, "decline")
}

// A life or damage price is paid above the Phyrexian life floor and
// declined below it (ADR 0150 §6, ADR 0131 §5).
func TestABotOpponentPaysLifeOrDamageOnlyAboveTheFloor(t *testing.T) {
	for _, damage := range []bool{false, true} {
		if pay, decline := lifeOfferValues(t, 40, 5, damage); pay <= decline {
			t.Errorf("damage=%v at 40 life: pay %.3f, decline %.3f; the bot must pay", damage, pay, decline)
		}
		if pay, decline := lifeOfferValues(t, 12, 5, damage); pay >= decline {
			t.Errorf("damage=%v at 12 life: pay %.3f, decline %.3f; 7 is below the floor", damage, pay, decline)
		}
	}
}

func TestIcyPrisonsOfferIsPaidByItsControllerAndNotByAnOpponent(t *testing.T) {
	if pay, decline := anyPlayerOfferValues(t, 0, 0, true); pay <= decline {
		t.Errorf("own Icy Prison: pay %.3f, decline %.3f; the controller keeps it", pay, decline)
	}
	if pay, decline := anyPlayerOfferValues(t, 1, 0, true); pay >= decline {
		t.Errorf("an opponent's Icy Prison: pay %.3f, decline %.3f; an opponent lets it go", pay, decline)
	}
}
