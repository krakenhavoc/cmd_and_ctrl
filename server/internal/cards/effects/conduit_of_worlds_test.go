package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// conduit_of_worlds_test.go — #2173: a cast permission with an "if you
// do" follow-up. Conduit of Worlds.

const conduitOfWorldsOracle = "ed14be15-8f8d-4fe3-a147-f5da8ed873bf"

// conduitTable sets up Conduit on seat 0's battlefield at main phase,
// a nonland permanent card and a land card in their graveyard, and a
// {R} instant in hand.
func conduitTable(t *testing.T) (g *game.Game, me *game.Player, conduit, bear, land, bolt uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	conduit = pushCatalogPermanent(g, me.ID, "Conduit of Worlds", "Artifact", conduitOfWorldsOracle, false)
	bear = pushGraveyardCardWithTypeLineID(me, "Grizzly Bears", "Creature — Bear")
	land = pushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
	bolt = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	// The bear needs a cost to be castable at all.
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == bear {
			me.Graveyard.Cards[i].ManaCost = "{1}{G}"
			me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 2, 2
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	return
}

func conduitActivate(t *testing.T, g *game.Game, me *game.Player, conduit, target uuid.UUID) error {
	t.Helper()
	err := g.ActivateCatalogAbility(me.ID, conduit, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err == nil {
		passPriorityAroundTable(t, g)
	}
	return err
}

func conduitCastFromGraveyard(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"})
}

func conduitMana(t *testing.T, g *game.Game, me *game.Player, cost string) {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, cost); err != nil {
		t.Fatal(err)
	}
}

func conduitBanned(err error) bool {
	var banned *game.CantCastError
	return errors.As(err, &banned)
}

func TestConduitOfWorldsCastsTheCardAndBansFurtherSpells(t *testing.T) {
	g, me, conduit, bear, _, bolt := conduitTable(t)
	conduitMana(t, g, me, "{C}{G}{R}")
	if err := conduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := conduitCastFromGraveyard(g, me, bear); err != nil {
		t.Fatalf("cast from graveyard: %v", err)
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); !conduitBanned(err) {
		t.Fatalf("a second spell this turn must be refused by the ban, got %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the spell cast through Conduit resolves")
	}
}

func TestConduitOfWorldsDeclinedOfferBansNothing(t *testing.T) {
	g, me, conduit, bear, _, bolt := conduitTable(t)
	conduitMana(t, g, me, "{C}{G}{R}")
	if err := conduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Declining is passing priority: the offer closes (CR 608.2g).
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if err := conduitCastFromGraveyard(g, me, bear); err == nil {
		t.Fatal("the offer closed on the pass; the card cannot be cast afterwards")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("declining the cast binds nobody: %v", err)
	}
}

func TestConduitOfWorldsExpiredOfferRunsNoFollowUp(t *testing.T) {
	g, me, conduit, bear, _, _ := conduitTable(t)
	if err := conduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The sandbox can move the turn without a pass; the offer lapses
	// with the turn and nothing follows.
	for i := 0; i < 40 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, s := range me.Statics {
		if s.CastBan.Kind != game.CastBanNone {
			t.Errorf("an unused offer must leave no ban behind: %+v", s)
		}
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("the expired offer is swept: %+v", me.CastPermissions)
	}
}

func TestConduitOfWorldsNoOfferAfterASpellWasCast(t *testing.T) {
	g, me, conduit, bear, _, bolt := conduitTable(t)
	conduitMana(t, g, me, "{C}{G}{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := conduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := conduitCastFromGraveyard(g, me, bear); err == nil {
		t.Error("\"if you haven't cast a spell this turn\": no offer after a spell was cast")
	}
}

func TestConduitOfWorldsOfferEndsWhenAnotherSpellIsCastFirst(t *testing.T) {
	g, me, conduit, bear, _, bolt := conduitTable(t)
	conduitMana(t, g, me, "{C}{G}{R}")
	if err := conduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the other spell first: %v", err)
	}
	if err := conduitCastFromGraveyard(g, me, bear); err == nil {
		t.Error("a spell cast first uses up the \"haven't cast a spell\" condition")
	}
}

func TestConduitOfWorldsTargetsOnlyNonlandPermanentCardsYouOwn(t *testing.T) {
	g, me, conduit, _, land, _ := conduitTable(t)
	spell := pushGraveyardCardWithTypeLineID(me, "Shock", "Instant")
	theirs := pushGraveyardCardWithTypeLineID(g.Seats[1], "Their Bear", "Creature — Bear")
	for _, id := range []uuid.UUID{land, spell, theirs} {
		if err := conduitActivate(t, g, me, conduit, id); err == nil {
			t.Errorf("%s is not a legal target", id)
		}
	}
}

func TestConduitOfWorldsPlaysLandsFromTheGraveyard(t *testing.T) {
	g, me, _, _, land, _ := conduitTable(t)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("the standing permission lets you play a land from the graveyard: %v", err)
	}
}

func TestConduitOfWorldsEnumeratorAgrees(t *testing.T) {
	g, me, conduit, bear, _, _ := conduitTable(t)
	conduitMana(t, g, me, "{C}{G}")
	offered := func() bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == legal.TypeCastSpell && m.Source == bear {
				return true
			}
		}
		return false
	}
	if offered() {
		t.Fatal("no cast offer before the ability resolves")
	}
	if err := conduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !offered() {
		t.Fatal("the enumerator offers the cast from the graveyard")
	}
	if err := conduitCastFromGraveyard(g, me, bear); err != nil {
		t.Fatalf("the engine accepts what the enumerator offered: %v", err)
	}
	if offered() {
		t.Error("the permission is spent by the cast")
	}
}
