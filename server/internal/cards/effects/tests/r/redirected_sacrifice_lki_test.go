package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// redirected_sacrifice_lki_test.go — #3859: a sacrificed card that a
// replacement sent somewhere other than a graveyard is still "that card"
// (Nazgûl Battle-Mace), and an effect that reads an object that has left
// its zone reads its last-known information (Skyclave Apparition's
// exiled card, Orvar's copied permanent; CR 608.2h).

// nazgulMaceTable equips a Battle-Mace on a Knight of mine and has the
// opponent sacrifice a Swamp, leaving the trigger on the stack.
func nazgulMaceTable(t *testing.T) (g *game.Game, me, opp *game.Player, swamp uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	host := effectstest.PushAnnCatalogCreature(g, me.ID, "Knight", "", 3, 3)
	mace := effectstest.PushCatalogPermanent(g, me.ID, "Nazgûl Battle-Mace", "Artifact — Equipment", effectstest.NazgulBattleMaceOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(mace, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	effectstest.AnnCard(t, g, host) // settle the layers so the granted trigger exists
	swamp = effectstest.PushCatalogPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", "", false)
	return g, me, opp, swamp
}

// Rest in Peace exiles the sacrificed card. It was sacrificed all the
// same, so the Mace still puts it onto the battlefield under its
// controller's control when the sacrificing player doesn't pay.
func TestNazgulBattleMaceTakesASacrificedCardThatRestInPeaceExiled(t *testing.T) {
	g, me, opp, swamp := nazgulMaceTable(t)
	effectstest.CastRestInPeace(t, g)

	if err := g.SacrificePermanent(opp.ID, swamp); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(swamp) {
		t.Fatal("setup: Rest in Peace did not exile the sacrificed card")
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // don't pay
	effectstest.PassPriorityAroundTable(t, g)

	// A card put onto the battlefield from exile is a new object (CR 400.7).
	if n := effectstest.CardsNamed(g, "Swamp", me.ID); n != 1 {
		t.Errorf("Swamps under my control = %d, want the exiled card, put onto the battlefield", n)
	}
}

// A shuffle-in or tuck replacement leaves the card in a library; the Mace
// finds it there too.
func TestNazgulBattleMaceTakesASacrificedCardThatWentToTheLibrary(t *testing.T) {
	g, me, opp, swamp := nazgulMaceTable(t)

	if err := g.SacrificePermanent(opp.ID, swamp); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.TuckToLibraryForEffect(swamp, false); err != nil {
			t.Fatalf("tuck: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // don't pay
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CardsNamed(g, "Swamp", me.ID); n != 1 {
		t.Errorf("Swamps under my control = %d, want the card from the library, put onto the battlefield", n)
	}
}

// Skyclave Apparition: the exiled card has since left exile (CR 608.2h).
// Its owner still gets the Illusion, as large as the card's mana value.
func TestSkyclaveApparitionPaysForAnExiledCardThatLeftExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID,
	})
	apparition := effectstest.CastSkyclaveApparition(t, g)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Fatal("the targeted permanent was not exiled")
	}

	g.WithWriteLock(func() {
		if err := g.PutIntoGraveyardForEffect(rock); err != nil {
			t.Fatalf("move the rock out of exile: %v", err)
		}
	})
	if g.Exile.Contains(rock) {
		t.Fatal("setup: the rock is still in exile")
	}
	effectstest.KillCreature(t, g, me.ID, apparition)
	effectstest.PassPriorityAroundTable(t, g)

	var illusion *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Illusion" {
			illusion = &g.Battlefield.Cards[i]
		}
	}
	if illusion == nil {
		t.Fatal("no Illusion token was created for a card that left exile")
	}
	if illusion.Controller != opp.ID || illusion.Power != 3 || illusion.Toughness != 3 {
		t.Errorf("the Illusion is %d/%d under %s, want 3/3 under the card's owner %s",
			illusion.Power, illusion.Toughness, illusion.Controller, opp.ID)
	}
}

// Orvar: a permanent that left the battlefield before the copy ability
// resolved is copied as it last existed there (CR 608.2h, 707.2).
func TestOrvarCopiesAPermanentThatLeftBeforeTheAbilityResolved(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, effectstest.ApaCreature("Orvar, the All-Form", effectstest.OrvarOracle, 3, 3))
	mine := effectstest.B12Creature(d.G, d.Caster.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, d.G, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	// The copy ability is on the stack above the spell; the Bear dies in
	// response.
	d.G.WithWriteLock(func() {
		if err := d.G.DestroyPermanentForEffect(mine); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, d.G)

	var token *game.Card
	for i := range d.G.Battlefield.Cards {
		c := &d.G.Battlefield.Cards[i]
		if c.Name == "My Bear" && c.InstanceID != mine {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no token copy of the Bear that died in response")
	}
	if !token.IsToken() || token.Controller != d.Caster.ID {
		t.Errorf("the copy is token=%v under %s, want a token under %s", token.IsToken(), token.Controller, d.Caster.ID)
	}
}

// Orvar: a permanent that blinked in response is a new object. The copy
// is of the one that was targeted (its last-known information), and the
// returned card is not copied twice.
func TestOrvarDoesNotMistakeABlinkedPermanentForTheTargetedOne(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, effectstest.ApaCreature("Orvar, the All-Form", effectstest.OrvarOracle, 3, 3))
	mine := effectstest.B12Creature(d.G, d.Caster.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, d.G, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	d.G.WithWriteLock(func() {
		if err := d.G.ExileCardForEffect(mine); err != nil {
			t.Fatalf("exile: %v", err)
		}
		if _, err := d.G.ReturnFromExileToBattlefieldForEffect(mine, d.Caster.ID, false); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, d.G)

	if n := d.CountNamedUnder(d.Caster, "My Bear"); n != 2 {
		t.Errorf("My Bears = %d, want the blinked Bear and one token copy", n)
	}
}
