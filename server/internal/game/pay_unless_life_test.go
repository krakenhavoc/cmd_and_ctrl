package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// pay_unless_life_test.go — ADR 0131 §2 (#2531), PR 2: a mana
// "unless that player pays {B}" (ward, Rhystic Study's tax) answered with
// 2 life per symbol, ResolvePayUnlessWithLife / `phyrexian_life` on the
// resolve_choice answer. A {B} under K'rrik, Son of Yawgmoth or a printed
// {B/P}.

// payUnlessLifeTable is seat 1 paying, holding K'rrik's grant when
// withGrant.
func payUnlessLifeTable(t *testing.T, withGrant bool) (*Game, *Player) {
	t.Helper()
	withLifeForManaStatic(t)
	g := newActiveGame(t)
	payer := g.Seats[1]
	if withGrant {
		pushKrrik(g, payer)
	}
	return g, payer
}

func pendingPayUnless(g *Game, id uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return true
		}
	}
	return false
}

// The headline: a ward {B} with an empty pool is paid with 2 life, and
// the decline consequence does not run.
func TestKrrikPaysAWardCostWithLife(t *testing.T) {
	g, payer := payUnlessLifeTable(t, true)
	life := payer.Life
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{B}", &declined)
	if err := g.ResolvePayUnlessWithLife(id, payer.ID, true, nil, 1); err != nil {
		t.Fatalf("{B} paid with 2 life: %v", err)
	}
	if declined != 0 {
		t.Error("the decline consequence ran after a payment")
	}
	if payer.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", payer.Life, life-PhyrexianLifePerSymbol)
	}
	if pendingPayUnless(g, id) {
		t.Error("the prompt was not dequeued")
	}
}

// One symbol by life, one by mana.
func TestKrrikMixesLifeAndManaOnAPayUnless(t *testing.T) {
	g, payer := payUnlessLifeTable(t, true)
	life := payer.Life
	payer.ManaPool.AddMana(ManaToken{Color: "B"})
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{B}{B}", &declined)
	if err := g.ResolvePayUnlessWithLife(id, payer.ID, true, nil, 1); err != nil {
		t.Fatalf("{B}{B} off one black mana and 2 life: %v", err)
	}
	if declined != 0 || payer.Life != life-PhyrexianLifePerSymbol || len(payer.ManaPool) != 0 {
		t.Errorf("declined %d, life %d (want %d), pool %v (want empty)", declined, payer.Life, life-PhyrexianLifePerSymbol, payer.ManaPool)
	}
}

// Without a claim the cost is as printed: an unfunded "yes" is the
// decline, and no life is spent.
func TestKrrikPayUnlessWithoutAClaimDeclines(t *testing.T) {
	g, payer := payUnlessLifeTable(t, true)
	life := payer.Life
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{B}", &declined)
	if err := g.ResolvePayUnless(id, payer.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 1 || payer.Life != life {
		t.Errorf("declined %d (want 1), life %d (want %d)", declined, payer.Life, life)
	}
}

// Auto-tap never pays life: a Swamp pays the {B}, the life stays.
func TestKrrikPayUnlessAutoTapNeverPaysLife(t *testing.T) {
	g, payer := payUnlessLifeTable(t, true)
	life := payer.Life
	swamp := pushBattlefieldForTest(g, payer.ID, "Swamp", "Basic Land — Swamp", "")
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{B}", &declined)
	if err := g.ResolvePayUnless(id, payer.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 0 || payer.Life != life || !tappedForTest(g, swamp) {
		t.Errorf("declined %d, life %d (want %d), swamp tapped %v", declined, payer.Life, life, tappedForTest(g, swamp))
	}
}

// A claim the cost cannot honour is REFUSED with the prompt left in
// place: nobody loses a ward payment to a client bug.
func TestPayUnlessLifeClaimsAreRefusedWithThePromptInPlace(t *testing.T) {
	for _, tc := range []struct {
		name      string
		withGrant bool
		cost      string
		apply     bool
		life      int
	}{
		{"no grant, no printed symbol", false, "{B}", true, 1},
		{"more symbols than the cost has", true, "{B}", true, 2},
		{"negative", true, "{B}", true, -1},
		{"generic mana is not reached", true, "{2}", true, 1},
		{"a decline cannot pay life", true, "{B}", false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, payer := payUnlessLifeTable(t, tc.withGrant)
			life := payer.Life
			var declined int
			id := queuePayUnless(t, g, payer.ID, tc.cost, &declined)
			err := g.ResolvePayUnlessWithLife(id, payer.ID, tc.apply, nil, tc.life)
			if !errors.Is(err, ErrInvalidParam) {
				t.Fatalf("got %v, want ErrInvalidParam", err)
			}
			if !pendingPayUnless(g, id) || declined != 0 || payer.Life != life {
				t.Errorf("a refused claim moved something: pending %v, declined %d, life %d",
					pendingPayUnless(g, id), declined, payer.Life)
			}
		})
	}
}

// CR 119.4 and CR 119.8: below 2 life, and under a life lock, the claim
// is refused with the prompt in place.
func TestKrrikPayUnlessClaimIsGatedByLife(t *testing.T) {
	g, payer := payUnlessLifeTable(t, true)
	payer.Life = 1
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{B}", &declined)
	if err := g.ResolvePayUnlessWithLife(id, payer.ID, true, nil, 1); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("2 life at 1 life: %v, want ErrInvalidParam", err)
	}
	if !pendingPayUnless(g, id) || payer.Life != 1 {
		t.Errorf("a refused claim moved something: pending %v, life %d", pendingPayUnless(g, id), payer.Life)
	}

	g, payer = payUnlessLifeTable(t, true)
	lockUntilNextTurn(g, payer)
	id = queuePayUnless(t, g, payer.ID, "{B}", &declined)
	if err := g.ResolvePayUnlessWithLife(id, payer.ID, true, nil, 1); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("life under a lock: %v, want ErrInvalidParam", err)
	}
	if !pendingPayUnless(g, id) {
		t.Error("a refused claim dequeued the prompt")
	}
}

// A well-formed claim whose remaining mana cannot be funded is the
// decline, with no life paid and no land tapped.
func TestKrrikUnfundedClaimDeclinesWithoutPayingLife(t *testing.T) {
	g, payer := payUnlessLifeTable(t, true)
	life := payer.Life
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{1}{B}", &declined)
	if err := g.ResolvePayUnlessWithLife(id, payer.ID, true, nil, 1); err != nil {
		t.Fatalf("ResolvePayUnlessWithLife: %v", err)
	}
	if declined != 1 || payer.Life != life {
		t.Errorf("declined %d (want 1), life %d (want %d)", declined, payer.Life, life)
	}
}

// A printed {B/P} in a pay-unless cost needs no grant.
func TestPrintedPhyrexianSymbolInAPayUnlessPaysWithLife(t *testing.T) {
	g, payer := payUnlessLifeTable(t, false)
	life := payer.Life
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{B/P}", &declined)
	if err := g.ResolvePayUnlessWithLife(id, payer.ID, true, nil, 1); err != nil {
		t.Fatalf("{B/P} paid with 2 life: %v", err)
	}
	if declined != 0 || payer.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("declined %d, life %d (want %d)", declined, payer.Life, life-PhyrexianLifePerSymbol)
	}
}
