package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_from_hand_test.go — ADR 0189 decision 3 (#3538): manifesting
// cards from a hand. Kozilek, the Broken Reality manifests two cards
// from each of up to two target players' hands; Scroll of Fate
// manifests one of yours.

const (
	KozilekBrokenRealityOracle = "39f4bf0e-47d5-4ae8-a504-71acda849850"
	ScrollOfFateOracle         = "c15c13d2-8db9-4801-a1f0-f087f2551c92"
)

// emptyHandForTest removes every card from a player's hand.
func EmptyHandForTest(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
}

// fillHandForTest gives a player a hand of n distinct cards and returns
// their IDs. Every third card is an instant, so a manifest of a
// nonpermanent is in the mix.
func FillHandForTest(g *game.Game, p *game.Player, n int) []uuid.UUID {
	EmptyHandForTest(g, p)
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		typeLine := "Creature — Bear"
		if i%3 == 2 {
			typeLine = "Instant"
		}
		ids = append(ids, HandCardForTest(p, "Hand card "+string(rune('A'+i)), typeLine, ""))
	}
	return ids
}

// castKozilek puts Kozilek in the active player's hand, casts it and
// answers the cast trigger's target prompt with `targets`. It leaves the
// trigger on the stack above the spell.
func CastKozilek(t *testing.T, g *game.Game, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := PutInHand(me, game.Card{Name: "Kozilek, the Broken Reality", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: KozilekBrokenRealityOracle, ManaCost: "{9}", Power: 9, Toughness: 9})
	ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the cast trigger did not ask for its target players")
	}
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	return id
}

// cardInPlay reports whether a card is on the battlefield.
func CardInPlay(g *game.Game, id uuid.UUID) bool {
	_, ok := BattlefieldCard(g, id)
	return ok
}

// assertManifestedBy checks a card is a face-down manifest under
// `owner`'s control that only `owner` can look at.
func AssertManifestedBy(t *testing.T, g *game.Game, id, owner uuid.UUID) {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	if !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownManifested {
		t.Fatalf("face-down state = (%v, %q), want a manifested permanent", c.FaceDown, c.FaceDownKind)
	}
	if c.Controller != owner {
		t.Errorf("controller = %s, want its owner %s", c.Controller, owner)
	}
	for _, p := range g.Seats {
		if known := c.IsKnownTo(p.ID); known != (p.ID == owner) {
			t.Errorf("seat %s knows the manifest = %v, want only its owner to (CR 708.5)", p.ID, known)
		}
	}
	if eff := c.Effective(); eff.Power != 2 || eff.Toughness != 2 {
		// The anthem tests assert the buffed body; here Kozilek is not
		// on the battlefield, so a manifest is the plain 2/2.
		t.Errorf("body = %d/%d, want a 2/2", eff.Power, eff.Toughness)
	}
}
