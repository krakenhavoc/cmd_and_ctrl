package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// either_payment_test.go — ADR 0187 §5 (#3575): the card side of an
// either payment. No card uses it yet (Worms of the Earth is slice 8), so
// the offer is applied straight from the effect.

func wormsEither() effects.UpkeepPayment {
	return effects.EitherPayment(
		effects.SacrificePayment(2, "land", "lands", effects.QueryType("land")),
		effects.DamagePayment(5),
	)
}

func TestEitherPaymentWordsAndBranches(t *testing.T) {
	p := wormsEither()
	if p.Words != "Sacrifice two lands or take 5 damage" {
		t.Errorf("words %q", p.Words)
	}
	if p.Action == nil || p.Action.Kind != game.PayActionEither || len(p.Action.Either) != 2 {
		t.Fatalf("action %+v, want an either of two", p.Action)
	}
	sac, dmg := p.Action.Either[0], p.Action.Either[1]
	if sac.Kind != game.PayActionSacrifice || sac.Count != 2 || sac.Words != "sacrifice two lands" {
		t.Errorf("branch 0 = %+v", sac)
	}
	if dmg.Kind != game.PayActionDamage || dmg.Count != 5 || dmg.Words != "take 5 damage" {
		t.Errorf("branch 1 = %+v", dmg)
	}
	if got := effects.EitherPayment(effects.LifePayment(3), effects.DiscardPayment(1)).Words; got != "Pay 3 life or discard a card" {
		t.Errorf("a life-or-discard either reads %q", got)
	}
}

// A mana payment, or an either inside an either, has no branch shape.
func TestEitherPaymentRefusesABranchItCannotHold(t *testing.T) {
	for name, build := range map[string]func(){
		"mana":   func() { effects.EitherPayment(effects.UpkeepPayment{Mana: "{1}"}, effects.DamagePayment(5)) },
		"nested": func() { effects.EitherPayment(wormsEither(), effects.DamagePayment(5)) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			build()
		}()
	}
}

// Through the card's own AnyPlayerPays: the prompt carries the whole
// offer's words, a player short of lands takes the damage branch, and the
// first payer ends the offer.
func TestEitherPaymentOfferAsksOnceAndTheAnswerNamesTheBranch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	first := g.Seats[0]
	effectstest.ResetAnyPlayerProbe()
	item := &game.StackItem{Controller: g.Seats[2].ID, SourceCardID: uuid.New()}
	g.WithWriteLock(func() {
		err := effects.AnyPlayerPays{
			Pay:      wormsEither(),
			Question: "Worms — sacrifice two lands or take 5 damage?",
			Then:     effectstest.TestAnyPlayerProbe,
		}.Apply(effects.NewContext(g, item))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	link := effectstest.AppOpenLink(g)
	if link == nil || link.Chooser != first.ID {
		t.Fatalf("open link %+v, want the active player's", link)
	}
	if link.PayCost != "sacrifice two lands or take 5 damage" {
		t.Errorf("link words %q", link.PayCost)
	}
	if a := link.PayAction(); a == nil || a.Kind != game.PayActionEither || len(a.Either) != 2 {
		t.Fatalf("link payment %+v", a)
	}
	one := 1
	if err := g.ResolvePayUnlessWithOption(link.ID, link.Chooser, true, &one, nil); err != nil {
		t.Fatalf("taking the damage: %v", err)
	}
	if first.Life != 35 {
		t.Errorf("life %d, want 40 - 5", first.Life)
	}
	paid, nobody, payer := effectstest.TestAnyPlayerPaid, effectstest.TestAnyPlayerNobody, effectstest.TestAnyPlayerPayer
	if paid != 1 || nobody != 0 || payer != first.ID || effectstest.AppOpenLink(g) != nil {
		t.Errorf("paid %d by %v, nobody %d, open link %v; want one payer, the first seat, and no more questions", paid, payer, nobody, effectstest.AppOpenLink(g))
	}
}
