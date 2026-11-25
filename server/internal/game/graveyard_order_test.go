package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// graveyard_order_test.go — #6194: the clauses that read where a card
// sits in a graveyard. CR 404.2: "Each graveyard is kept in a single
// face-up pile. A player can examine the cards in any graveyard at any
// time but normally can't change their order."

// pileOf seeds p's graveyard bottom to top with cards of the given type
// lines and returns their IDs in that order.
func pileOf(p *Player, typeLines ...string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(typeLines))
	for _, tl := range typeLines {
		out = append(out, seedGraveyard(p, "Pile "+tl, tl, "{1}"))
	}
	return out
}

func isCreatureCard(c Card) bool { return c.HasCardType("creature") }

// Zombie Scavengers' "Exile the top creature card of your graveyard": the
// cost offers the topmost creature card and no other, and refuses a
// creature card below it.
func TestTopOnlyExileCostTakesTheTopmostMatchAndNoOther(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := pileOf(me, "Creature — Bear", "Creature — Elf", "Sorcery")
	low, top := ids[0], ids[1]
	cost := &ExileCost{N: 1, Label: "the top creature card", From: ZoneGraveyard, Match: isCreatureCard, TopOnly: true}

	g.WithWriteLock(func() {
		if got := g.ExileCostOptionsForEffect(me.ID, uuid.Nil, cost); !slices.Equal(got, []uuid.UUID{top}) {
			t.Fatalf("options are %v, want only the topmost creature card %v", got, top)
		}
		if _, err := g.validateExileCardsCostLocked(me.ID, uuid.Nil, cost, []uuid.UUID{low}, nil); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a creature card below the top one paid: %v", err)
		}
		if _, err := g.validateExileCardsCostLocked(me.ID, uuid.Nil, cost, []uuid.UUID{ids[2]}, nil); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("the top card, a sorcery, paid a creature clause: %v", err)
		}
		if _, err := g.validateExileCardsCostLocked(me.ID, uuid.Nil, cost, []uuid.UUID{top}, nil); err != nil {
			t.Errorf("the topmost creature card did not pay: %v", err)
		}
	})

	// With no creature card at all, there is nothing to pay with (CR 118.3).
	g2 := newActiveGame(t)
	pileOf(g2.Seats[0], "Sorcery", "Instant")
	g2.WithWriteLock(func() {
		if g2.ExileCostPayableForEffect(g2.Seats[0].ID, uuid.Nil, cost) {
			t.Error("a graveyard with no creature card can pay the clause")
		}
	})
}

// Barrow Ghoul's "sacrifice this creature unless you exile the top
// creature card of your graveyard": the payment's options are that one
// card, and a pick of another is refused.
func TestTopOnlyGraveyardPaymentOffersTheTopmostCreatureCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := pileOf(me, "Creature — Bear", "Creature — Elf", "Instant")
	a := &PayAction{Kind: PayActionExileFromGraveyard, Count: 1, CardType: "creature", TopOnly: true}
	g.WithWriteLock(func() {
		if got := g.PayActionOptionsForEffect(me.ID, a); !slices.Equal(got, []uuid.UUID{ids[1]}) {
			t.Fatalf("options are %v, want the topmost creature card", got)
		}
		if err := g.validatePayActionLocked(me.ID, a, []uuid.UUID{ids[0]}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a creature card below the top one paid: %v", err)
		}
		if err := g.validatePayActionLocked(me.ID, a, []uuid.UUID{ids[1]}); err != nil {
			t.Errorf("the topmost creature card did not pay: %v", err)
		}
	})
}

// Spinning Darkness's "exile the top three black cards of your graveyard
// rather than pay this spell's mana cost": the three topmost black cards
// are the candidates, and a black card below them is refused.
func TestGraveyardTopOnlyAlternativeCostNamesTheTopmostMatches(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	for _, cost := range []string{"{B}", "{B}", "{U}", "{B}", "{B}"} {
		ids = append(ids, seedGraveyard(me, "Card "+cost, "Sorcery", cost))
	}
	black := &TargetSpec{CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
		return c.ColorSymbols("B") > 0
	}, Min: 3, Max: 3}
	alt := &AlternativeCost{Key: "test-top-three", ExileFromGraveyard: black, GraveyardTopOnly: true}
	g.WithWriteLock(func() {
		got := g.AltCostCandidatesLocked(me.ID, uuid.Nil, alt)
		want := []uuid.UUID{ids[4], ids[3], ids[1]}
		if !slices.Equal(got, want) {
			t.Fatalf("candidates are %v, want the three topmost black cards %v", got, want)
		}
		if err := g.validateAlternativeCostPaymentLocked(me.ID, uuid.Nil, alt, []uuid.UUID{ids[4], ids[3], ids[0]}, 0); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("a black card below the top three paid: %v", err)
		}
		if err := g.validateAlternativeCostPaymentLocked(me.ID, uuid.Nil, alt, want, 0); err != nil {
			t.Errorf("the top three black cards did not pay: %v", err)
		}
	})
}

// Bösium Strip's "you may cast instant and sorcery spells from the top of
// your graveyard": the top card is opened, an instant below it is not.
func TestTopOfGraveyardPermissionOpensOnlyTheTopCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	below := seedGraveyard(me, "Below", "Instant", "{U}")
	top := seedGraveyard(me, "Top", "Instant", "{U}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, Scope: ScopeStanding,
			Filter: PermissionFilter{InstantOrSorceryOnly: true}, Duration: g.UntilEndOfTurnDuration(),
			TopOfGraveyardOnly: true, ExileInsteadOfGraveyard: true,
		})
	})
	if perm := grantOn(g, me.ID, top, ZoneGraveyard); !perm.Granted() {
		t.Fatal("the top instant of the graveyard is not castable")
	}
	if perm := grantOn(g, me.ID, below, ZoneGraveyard); perm.Granted() {
		t.Error("an instant below the top of the graveyard is castable")
	}
	// A sorcery put on top covers what was the top card.
	seedGraveyard(me, "Land", "Land", "")
	if perm := grantOn(g, me.ID, top, ZoneGraveyard); perm.Granted() {
		t.Error("an instant that is no longer on top is still castable")
	}
}

// Krovikan Horror's "with a creature card directly above it" and Ashen
// Ghoul's "three or more creature cards above this card".
func TestGraveyardPositionReaders(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := pileOf(me, "Creature — Horror", "Creature — Bear", "Sorcery", "Creature — Elf", "Creature — Wolf")
	horror := ids[0]
	g.WithWriteLock(func() {
		if !g.CardDirectlyAboveIsForEffect(horror, isCreatureCard) {
			t.Error("a creature card directly above was not seen")
		}
		if g.CardDirectlyAboveIsForEffect(ids[1], isCreatureCard) {
			t.Error("a sorcery directly above read as a creature card")
		}
		if g.CardDirectlyAboveIsForEffect(ids[4], nil) {
			t.Error("the top card has a card above it")
		}
		if n := g.CountGraveyardCardsAboveForEffect(horror, isCreatureCard); n != 3 {
			t.Errorf("creature cards above the bottom card: %d, want 3", n)
		}
		if got := g.BottomGraveyardCardForEffect(me.ID); got != horror {
			t.Errorf("bottom card is %v, want %v", got, horror)
		}
		if got := g.TopGraveyardCardsForEffect(me.ID, 2, isCreatureCard); !slices.Equal(got, []uuid.UUID{ids[4], ids[3]}) {
			t.Errorf("top two creature cards are %v", got)
		}
	})
}

// "Then reorder your graveyard as you choose", and at random: the order
// is the pile's, and a snapshot keeps it (CR 404.2).
func TestReorderGraveyardSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := pileOf(me, "Sorcery", "Instant", "Creature — Bear")
	g.WithWriteLock(func() {
		if err := g.ReorderGraveyardForEffect(me.ID, []uuid.UUID{ids[0], ids[2]}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("an order missing a card was taken: %v", err)
		}
		if err := g.ReorderGraveyardForEffect(me.ID, []uuid.UUID{ids[0], ids[0], ids[2]}); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("an order naming a card twice was taken: %v", err)
		}
		// Top first: the sorcery on top, the bear at the bottom.
		if err := g.ReorderGraveyardForEffect(me.ID, []uuid.UUID{ids[0], ids[1], ids[2]}); err != nil {
			t.Fatalf("reorder: %v", err)
		}
	})
	want := []uuid.UUID{ids[2], ids[1], ids[0]}
	_, restored := roundTrip(t, g)
	var got []uuid.UUID
	for _, c := range restored.Seats[0].Graveyard.Cards {
		got = append(got, c.InstanceID)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("restored pile is %v, want %v (bottom to top)", got, want)
	}

	// At random: the same cards, from the seat's stream.
	g.WithWriteLock(func() { g.ReorderGraveyardAtRandomForEffect(me.ID) })
	var after []uuid.UUID
	for _, c := range me.Graveyard.Cards {
		after = append(after, c.InstanceID)
	}
	slices.SortFunc(after, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	slices.SortFunc(want, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	if !slices.Equal(after, want) {
		t.Errorf("a random reorder changed the cards: %v", after)
	}
}
