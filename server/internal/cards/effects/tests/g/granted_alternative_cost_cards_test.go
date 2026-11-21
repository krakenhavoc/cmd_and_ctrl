package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestWUBRGGrantorsOfferAndCharge(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Fist of Suns", effectstest.FistOfSunsOracle, "Artifact"},
		{"Jodah, Archmage Eternal", effectstest.JodahArchmageOracle, "Legendary Creature — Human Wizard"},
		{"Leyline of Mutation", effectstest.LeylineOfMutationOracle, "Enchantment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.ToMain(t, g)
			src := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine,
				OracleID: tc.oracle, Owner: me.ID, Controller: me.ID,
			})
			spell := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: spell, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
				Owner: me.ID, Controller: me.ID,
			})

			got := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, spell, game.ZoneHand), effects.GrantedAltCostWUBRG)
			if got == nil {
				t.Fatalf("no %s offer from hand with %s out", effects.GrantedAltCostWUBRG, tc.name)
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
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostWUBRG}); err != nil {
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
			if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, other, game.ZoneHand), effects.GrantedAltCostWUBRG) != nil {
				t.Errorf("the offer outlived %s", tc.name)
			}
		})
	}
}

func TestOmniscienceOffersAFreeCastFromHandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	src := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omniscience", TypeLine: "Enchantment",
		OracleID: effectstest.OmniscienceOracle, Owner: me.ID, Controller: me.ID,
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

	got := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, spell, game.ZoneHand), effects.GrantedAltCostFree)
	if got == nil || got.ManaCost != "" || !got.Granted {
		t.Fatalf("free offer from hand = %+v", got)
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, inYard, game.ZoneGraveyard), effects.GrantedAltCostFree) != nil {
		t.Errorf("Omniscience offered a free cast from a graveyard")
	}

	// Empty pool, strict: the {6}{U} is not paid.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFree}); err != nil {
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
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, other, game.ZoneHand), effects.GrantedAltCostFree) != nil {
		t.Errorf("the free offer outlived Omniscience")
	}
}
