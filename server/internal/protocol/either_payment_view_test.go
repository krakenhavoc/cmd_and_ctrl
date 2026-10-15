package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// either_payment_view_test.go — ADR 0187 §5, §8 (#3575): an either
// payment ships `pay_either`, one entry per branch, and none of the other
// payment fields.

func queueEitherView(t *testing.T, g *game.Game, branches ...game.PayAction) {
	t.Helper()
	controller := g.Seats[1].ID
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Worms", TypeLine: "Enchantment", Owner: controller, Controller: controller})
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
			Source: source, Controller: controller, Cost: "sacrifice two lands or take 5 damage", Question: "Worms?",
			Action: &game.PayAction{Kind: game.PayActionEither, Either: branches},
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})
}

func wormsBranches() []game.PayAction {
	return []game.PayAction{
		{Kind: game.PayActionSacrifice, Count: 2, Of: game.PermanentQuery{Types: []string{"land"}}, Words: "sacrifice two lands"},
		{Kind: game.PayActionDamage, Count: 5, Words: "take 5 damage"},
	}
}

// Each branch is an entry: the sacrifice branch with the chooser's own
// lands as options (public, so every seat gets them), the damage branch
// with none. No other payment field rides beside it.
func TestEitherPaymentIsOnTheWireAsOneEntryPerBranch(t *testing.T) {
	g := buildActiveGame(t)
	queueEitherView(t, g, wormsBranches()...)
	chooser := g.PendingChoices[0].Chooser
	var lands []string
	for _, s := range g.Seats {
		if s.ID != chooser {
			continue
		}
		for i := 0; i < 3; i++ {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: s.ID, Controller: s.ID})
			lands = append(lands, id.String())
		}
	}
	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String()).PendingChoices[0]
		if len(v.PayEither) != 2 {
			t.Fatalf("viewer %s: %d branches, want 2", viewer.Name, len(v.PayEither))
		}
		if v.PayCards != nil || v.PayDamage != nil || v.PayLife != nil || v.PayEnergy != nil || v.PayNothing {
			t.Errorf("viewer %s: an either also sent another payment: %+v", viewer.Name, v)
		}
		sac, dmg := v.PayEither[0], v.PayEither[1]
		if sac.Action != "sacrifice" || sac.Count != 2 || sac.Words != "sacrifice two lands" || !sac.Payable || len(sac.Options) != len(lands) {
			t.Errorf("viewer %s: branch 0 = %+v, want a payable two-land sacrifice with %d options", viewer.Name, sac, len(lands))
		}
		if dmg.Action != "damage" || dmg.Count != 5 || dmg.Words != "take 5 damage" || !dmg.Payable || len(dmg.Options) != 0 {
			t.Errorf("viewer %s: branch 1 = %+v", viewer.Name, dmg)
		}
	}
	raw, err := json.Marshal(ViewOfGameFor(g, g.Seats[0].ID.String()).PendingChoices[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"pay_either":[`) || strings.Contains(string(raw), `"pay_cards"`) {
		t.Errorf("wire form %s", raw)
	}
}

// A branch the chooser cannot pay is marked, not dropped: the client
// greys it with the reason.
func TestEitherPaymentMarksAnUnpayableBranch(t *testing.T) {
	g := buildActiveGame(t)
	queueEitherView(t, g, wormsBranches()...)
	v := ViewOfGameFor(g, g.PendingChoices[0].Chooser.String()).PendingChoices[0]
	if v.PayEither[0].Payable {
		t.Error("no lands, yet 'sacrifice two lands' is payable")
	}
	if !v.PayEither[1].Payable {
		t.Error("the damage branch is always payable")
	}
}

// A discard branch's options are the chooser's hand: only the chooser
// gets them.
func TestEitherPaymentDiscardOptionsGoToTheChooserOnly(t *testing.T) {
	g := buildActiveGame(t)
	queueEitherView(t, g,
		game.PayAction{Kind: game.PayActionDiscard, Count: 1, Words: "discard a card"},
		game.PayAction{Kind: game.PayActionDamage, Count: 5, Words: "take 5 damage"})
	chooser := g.PendingChoices[0].Chooser
	for _, viewer := range g.Seats {
		v := FilterViewFor(ViewOfGame(g), viewer.ID.String()).PendingChoices[0]
		got := len(v.PayEither[0].Options)
		if viewer.ID == chooser && got == 0 {
			t.Error("the chooser was not shown their own hand to discard from")
		}
		if viewer.ID != chooser && got != 0 {
			t.Errorf("viewer %s was shown %d of the chooser's hand cards", viewer.Name, got)
		}
	}
}
