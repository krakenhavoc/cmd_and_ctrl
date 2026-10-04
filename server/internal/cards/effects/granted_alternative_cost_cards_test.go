package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_alternative_cost_cards_test.go — ADR 0118 §3, PR 6 (#2163):
// the four cards that declare the granted-alternative-cost statics,
// each through the real catalog and the real cast path. Per card: the
// offer appears in the picker's list with the source's name in its
// label, claiming it pays the granted price under the strict gate, and
// the offer goes when the source leaves.

const (
	fistOfSunsOracle        = "6fd5e591-fae8-4128-a4d3-a848c8a8ffda"
	jodahArchmageOracle     = "8be4745e-36d8-430f-945e-c8a7fde9b4f6"
	leylineOfMutationOracle = "caab67eb-65e7-4755-b116-6977e97f0844"
	omniscienceOracle       = "730e39e6-c61d-48b5-8827-bfd952bf1be7"
)

func grantedCardOffers(g *game.Game, seat uuid.UUID, card uuid.UUID, zone game.ZoneKind) []*game.AlternativeCost {
	var out []*game.AlternativeCost
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		out = g.CastOffersForLocked(seat, c, zone, g.CastPermissionForLocked(seat, c, zone))
	})
	return out
}

func offerWithKey(offers []*game.AlternativeCost, key string) *game.AlternativeCost {
	for _, o := range offers {
		if o != nil && o.Key == key {
			return o
		}
	}
	return nil
}

func TestWUBRGGrantorsOfferAndCharge(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Fist of Suns", fistOfSunsOracle, "Artifact"},
		{"Jodah, Archmage Eternal", jodahArchmageOracle, "Legendary Creature — Human Wizard"},
		{"Leyline of Mutation", leylineOfMutationOracle, "Enchantment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			toMain(t, g)
			src := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine,
				OracleID: tc.oracle, Owner: me.ID, Controller: me.ID,
			})
			spell := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: spell, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
				Owner: me.ID, Controller: me.ID,
			})

			got := offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), GrantedAltCostWUBRG)
			if got == nil {
				t.Fatalf("no %s offer from hand with %s out", GrantedAltCostWUBRG, tc.name)
			}
			if got.ManaCost != "{W}{U}{B}{R}{G}" || !got.Granted {
				t.Errorf("offer = %+v", got)
			}
			if want := "(" + tc.name + ")"; len(got.Label) < len(want) || got.Label[len(got.Label)-len(want):] != want {
				t.Errorf("label %q does not end with the source's name", got.Label)
			}

			// Strict, with exactly {W}{U}{B}{R}{G} floating: the printed
			// {6}{U} could not be paid, so this proves the swap.
			me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "B"},
				game.ManaToken{Color: "R"}, game.ManaToken{Color: "G"})
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostWUBRG}); err != nil {
				t.Fatalf("cast through %s: %v", tc.name, err)
			}
			if me.Hand.Contains(spell) {
				t.Errorf("the spell is still in hand")
			}

			// The source leaves: the offer goes with it.
			other := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: other, Name: "Other Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
				Owner: me.ID, Controller: me.ID,
			})
			g.WithWriteLock(func() {
				if _, err := game.MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
					t.Fatalf("remove the source: %v", err)
				}
			})
			if offerWithKey(grantedCardOffers(g, me.ID, other, game.ZoneHand), GrantedAltCostWUBRG) != nil {
				t.Errorf("the offer outlived %s", tc.name)
			}
		})
	}
}

func TestOmniscienceOffersAFreeCastFromHandOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	src := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omniscience", TypeLine: "Enchantment",
		OracleID: omniscienceOracle, Owner: me.ID, Controller: me.ID,
	})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	inYard := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: inYard, Name: "Yard Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
		Owner: me.ID, Controller: me.ID,
	})

	got := offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), GrantedAltCostFree)
	if got == nil || got.ManaCost != "" || !got.Granted {
		t.Fatalf("free offer from hand = %+v", got)
	}
	if offerWithKey(grantedCardOffers(g, me.ID, inYard, game.ZoneGraveyard), GrantedAltCostFree) != nil {
		t.Errorf("Omniscience offered a free cast from a graveyard")
	}

	// Empty pool, strict: the {6}{U} is not paid.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree}); err != nil {
		t.Fatalf("free cast through Omniscience: %v", err)
	}
	if me.Hand.Contains(spell) {
		t.Errorf("the spell is still in hand")
	}

	other := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: other, Name: "Other Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			t.Fatalf("remove the source: %v", err)
		}
	})
	if offerWithKey(grantedCardOffers(g, me.ID, other, game.ZoneHand), GrantedAltCostFree) != nil {
		t.Errorf("the free offer outlived Omniscience")
	}
}
