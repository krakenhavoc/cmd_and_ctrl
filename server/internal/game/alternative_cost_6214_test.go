package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// alternative_cost_6214_test.go — #6214: three alternative costs (CR
// 118.9: "An alternative cost is a cost listed in a spell's text, or
// applied to it from another effect, that its controller may pay rather
// than paying the spell's mana cost").
//
// Land Grant reveals the caster's hand, Reverent Silence and Skyshroud
// Cutter have each OTHER player gain life, and Dream Halls grants every
// player "discard a card that shares a color with that spell".

// altCost6214HandCard puts a card of `colors` in p's hand and returns its id.
func altCost6214HandCard(p *Player, name, typeLine string, colors ...string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine, c.Colors = typeLine, colors
	p.Hand.PushTop(c)
	return c.InstanceID
}

// Land Grant: with no land in hand the offer is there, costs no mana,
// and paying it shows the hand to the table (the Land Grant itself has
// already left it, CR 601.2a).
func TestRevealHandAlternativeCostShowsTheHand(t *testing.T) {
	const oracle = "test-land-grant"
	noLandsInHand := func(g *Game, controller uuid.UUID) bool {
		p := g.playerByIDLocked(controller)
		if p == nil {
			return false
		}
		for _, c := range p.Hand.Cards {
			if c.IsLand() {
				return false
			}
		}
		return true
	}
	offer := AlternativeCost{
		Key: "land-grant", Label: "Reveal your hand rather than pay this spell's mana cost",
		RevealHand: true, Condition: noLandsInHand,
	}

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	withCatalogAlternativeCosts(t, altCostFor(oracle, offer))
	id := blastInHand(t, g, me, oracle)
	bear := altCost6214HandCard(me, "Bear", "Creature — Bear", "G")

	var payable bool
	g.WithWriteLock(func() { payable = g.AlternativeCostPayableLocked(me.ID, id, &offer) })
	if !payable {
		t.Fatal("a hand with no land cannot pay a reveal-your-hand cost")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "land-grant"}); err != nil {
		t.Fatalf("cast for the alternative cost: %v", err)
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.IsKnownTo(opp.ID) {
		t.Error("the hand was not revealed to the other seats")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want nothing spent", me.ManaPool)
	}

	// A land in hand turns the offer off (the card's own condition).
	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	withCatalogAlternativeCosts(t, altCostFor(oracle, offer))
	id2 := blastInHand(t, g2, me2, oracle)
	altCost6214HandCard(me2, "Forest", "Basic Land — Forest")
	g2.WithWriteLock(func() { payable = g2.AlternativeCostPayableLocked(me2.ID, id2, &offer) })
	if payable {
		t.Error("a hand holding a land can pay Land Grant's cost")
	}
	if err := g2.CastSpell(me2.ID, id2, CastSpellParams{Strict: true, AlternativeCost: "land-grant"}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("claiming it with a land in hand: %v, want ErrInvalidParam", err)
	}
}

// Reverent Silence: each OTHER player gains 6 life, the caster none,
// and the cast costs no mana.
func TestEachOtherPlayerGainsLifeAlternativeCost(t *testing.T) {
	const oracle = "test-reverent-silence"
	offer := AlternativeCost{
		Key: "forest-silence", Label: "Have each other player gain 6 life rather than pay this spell's mana cost",
		EachOtherPlayerGainsLife: 6,
	}
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	withCatalogAlternativeCosts(t, altCostFor(oracle, offer))
	id := blastInHand(t, g, me, oracle)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "forest-silence"}); err != nil {
		t.Fatalf("cast for the alternative cost: %v", err)
	}
	for i, p := range g.Seats {
		want := before[i] + 6
		if p.ID == me.ID {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want nothing spent", me.ManaPool)
	}
}

// Dream Halls: a granted cost every player may pay, and only with a
// card that shares a colour with the spell being cast.
func TestGrantedDiscardSharesAColourWithTheSpell(t *testing.T) {
	const (
		halls  = "test-dream-halls"
		oracle = "test-blue-spell"
		key    = "granted-dream-halls"
	)
	granted := GrantedAlternativeCost{
		Offer: AlternativeCost{
			Key: key, Label: "Discard a card that shares a color with that spell",
			DiscardFromHand: &TargetSpec{}, DiscardSharesColorWithSpell: true,
		},
		Zones:       []ZoneKind{ZoneHand},
		EveryPlayer: true,
	}
	prev := CatalogGrantedAlternativeCosts
	CatalogGrantedAlternativeCosts = func(k string) []GrantedAlternativeCost {
		if k == halls {
			return []GrantedAlternativeCost{granted}
		}
		return nil
	}
	t.Cleanup(func() { CatalogGrantedAlternativeCosts = prev })

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	modifierSource(t, g, opp, "Test Dream Halls", halls)
	advanceTo(t, g, StepPrecombatMain)
	spell := NewCard("Test Counterspell", me.ID)
	spell.TypeLine, spell.ManaCost, spell.Colors, spell.OracleID = "Instant", "{U}{U}", []string{"U"}, oracle
	me.Hand.PushTop(spell)
	red := altCost6214HandCard(me, "Red Card", "Instant", "R")
	blue := altCost6214HandCard(me, "Blue Card", "Instant", "U")

	// The offer reaches a player who does not control the grantor.
	if keys := offerKeys(t, g, me, spell.InstanceID, ZoneHand); !hasKey(keys, key) {
		t.Fatalf("offers = %v, want the granted cost (its text says 'its controller may')", keys)
	}
	var candidates []uuid.UUID
	g.WithWriteLock(func() {
		candidates = g.AltCostCandidatesLocked(me.ID, spell.InstanceID, grantedOffer(t, g, me, spell.InstanceID, ZoneHand, key))
	})
	if len(candidates) != 1 || candidates[0] != blue {
		t.Fatalf("candidates = %v, want only the blue card (it must share a colour with the spell)", candidates)
	}
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Strict: true, AlternativeCost: key, AltCostIDs: []uuid.UUID{red},
	}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("discarding a red card for a blue spell: %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Strict: true, AlternativeCost: key, AltCostIDs: []uuid.UUID{blue},
	}); err != nil {
		t.Fatalf("discard the blue card: %v", err)
	}
	if me.Hand.Contains(blue) || !me.Graveyard.Contains(blue) {
		t.Error("the blue card was not discarded as the cost")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want nothing spent", me.ManaPool)
	}
}
