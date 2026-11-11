package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const extractBrainOracle = "bb0eb9c2-c26d-452d-9e02-fc0bec2655c4"

// extractBrainHand replaces a player's hand with one card per type line
// (cost {2}{G}) and returns their ids in order.
func extractBrainHand(p *game.Player, typeLines ...string) []uuid.UUID {
	p.Hand.Cards = nil
	ids := make([]uuid.UUID, 0, len(typeLines))
	for i, tl := range typeLines {
		c := game.NewCard("Their Card "+string(rune('A'+i)), p.ID)
		c.TypeLine = tl
		c.ManaCost = "{2}{G}"
		c.Power, c.Toughness = 2, 2
		c.AddKnower(p.ID)
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// extractBrainCast casts Extract Brain with X at opp and passes priority
// until it starts to resolve.
func extractBrainCast(t *testing.T, g *game.Game, opp uuid.UUID, x int) {
	t.Helper()
	b12PlayFromHand(t, g, "Extract Brain", "Sorcery", extractBrainOracle,
		game.CastSpellParams{XValue: x, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp}}})
	passPriorityAroundTable(t, g)
}

func extractBrainKnows(t *testing.T, g *game.Game, id, player uuid.UUID) bool {
	t.Helper()
	c := e2Card(t, g, id)
	return c.IsKnownTo(player)
}

// The opponent chooses X cards; only those are looked at, only by the
// caster, and one of them may be cast free out of the opponent's hand.
func TestExtractBrainCastsOneChosenCardFromTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	hand := extractBrainHand(opp, "Creature — Bear", "Creature — Elk", "Land", "Creature — Ox")
	extractBrainCast(t, g, opp.ID, 2)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("want the opponent asked for exactly two cards, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, hand[:1]); err == nil {
		t.Fatal("one card was accepted for X=2")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{hand[0], hand[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for i, id := range hand {
		chosen := i == 0 || i == 2
		if got := extractBrainKnows(t, g, id, me.ID); got != chosen {
			t.Errorf("card %d: the caster knows it = %v, want %v", i, got, chosen)
		}
		if extractBrainKnows(t, g, id, other.ID) {
			t.Errorf("card %d: a third player was shown it", i)
		}
	}

	offer := latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID)
	if offer == nil {
		t.Fatalf("no resolution_cast offer for the caster: %+v", g.PendingChoices)
	}
	// The land is looked at but is not a spell; the creature is on offer.
	if len(offer.DiscardOptions) != 1 || offer.DiscardOptions[0] != hand[0] {
		t.Fatalf("offer = %v, want only the chosen creature card", offer.DiscardOptions)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, hand[1], game.CastSpellParams{FromZone: "hand"}); err == nil {
		t.Fatal("a card the opponent did not choose was cast")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, hand[0], game.CastSpellParams{FromZone: "hand"}); err != nil {
		t.Fatalf("cast the chosen creature out of their hand for free: %v", err)
	}
	if got := b12ZoneOf(g, hand[0]); got != game.ZoneStack {
		t.Fatalf("the cast card is in %q, want the stack", got)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, hand[0])
	if c == nil || c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("the stolen creature: %+v, want it on the battlefield under the caster's control, owned by the opponent", c)
	}
	for _, id := range hand[1:] {
		if got := b12ZoneOf(g, id); got != game.ZoneHand {
			t.Errorf("an uncast card is in %q, want their hand", got)
		}
	}
}

// Declining leaves every card in the opponent's hand and closes the
// window: nothing in their hand is castable afterwards.
func TestExtractBrainDeclineLeavesTheCardsAndClosesTheWindow(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	hand := extractBrainHand(opp, "Creature — Bear", "Instant")
	// X covers the whole hand: nothing to choose.
	extractBrainCast(t, g, opp.ID, 5)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID); pick != nil {
		t.Fatalf("a hand of two was asked to choose for X=5: %+v", pick)
	}
	offer := latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID)
	if offer == nil || len(offer.DiscardOptions) != 2 {
		t.Fatalf("want both cards on offer, got %+v", offer)
	}
	// The moves the caster is offered are casts out of a hand that claim
	// the offer, and the decline.
	casts := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.ResolutionCast != offer.ID.String() || m.Type != legal.TypeCastSpell {
			continue
		}
		casts++
		var p struct {
			FromZone        string `json:"from_zone"`
			AlternativeCost string `json:"alternative_cost"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		if p.FromZone != "hand" || p.AlternativeCost == "" {
			t.Errorf("move %s: want a cast from a hand that claims the free offer", m.Params)
		}
	}
	if casts == 0 {
		t.Fatal("no cast of the looked-at cards was offered")
	}
	if err := g.ResolvePendingChoice(offer.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	for _, id := range hand {
		if got := b12ZoneOf(g, id); got != game.ZoneHand {
			t.Errorf("a declined card is in %q, want their hand", got)
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "hand", Strict: true, AutoTap: true}); err == nil {
			t.Error("a card in the opponent's hand was cast after the offer closed")
		}
	}
}

// A card in another player's hand is cast through the permission that
// names it and no other way: with a live grant over it, a cast that claims
// nothing (the printed cost, which a hand's owner needs no permission for)
// is refused, and the claimed cast is free.
func TestACardInAnotherHandIsCastOnlyUnderItsPermission(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	hand := extractBrainHand(opp, "Creature — Bear", "Creature — Elk")
	var lands []uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < 3; i++ {
			id := uuid.New()
			lands = append(lands, id)
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest",
				Owner: me.ID, Controller: me.ID,
			})
		}
		if !g.FreeCastFromHandForEffect(me.ID, uuid.Nil, hand[0], "Test Grant") {
			t.Fatal("the grant found no card")
		}
	})
	paid := game.CastSpellParams{FromZone: "hand", Strict: true, AutoTap: true}
	if err := g.CastSpell(me.ID, hand[1], paid); err == nil {
		t.Fatal("a card in the opponent's hand that no permission names was cast")
	}
	if err := g.CastSpell(me.ID, hand[0], paid); !errors.Is(err, game.ErrCardNotFound) {
		t.Fatalf("an unclaimed cast out of the opponent's hand: err = %v, want ErrCardNotFound", err)
	}
	assertTapped(t, g, false, lands...)
	free := paid
	free.AlternativeCost = game.AltCostKeyGrantedCost
	if err := g.CastSpell(me.ID, hand[0], free); err != nil {
		t.Fatalf("the claimed free cast: %v", err)
	}
	assertTapped(t, g, false, lands...)
	if got := b12ZoneOf(g, hand[0]); got != game.ZoneStack {
		t.Fatalf("the cast card is in %q, want the stack", got)
	}
}

// X=0 and an empty hand do nothing.
func TestExtractBrainWithNothingToChoose(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	hand := extractBrainHand(opp, "Creature — Bear")
	extractBrainCast(t, g, opp.ID, 0)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID); c != nil {
		t.Fatalf("X=0 offered a cast: %+v", c)
	}
	if extractBrainKnows(t, g, hand[0], me.ID) {
		t.Error("X=0 looked at a card")
	}
}
