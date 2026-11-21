package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func conduitCastFromGraveyard(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"})
}

func conduitBanned(err error) bool {
	var banned *game.CantCastError
	return errors.As(err, &banned)
}

func TestConduitOfWorldsCastsTheCardAndBansFurtherSpells(t *testing.T) {
	g, me, conduit, bear, _, bolt := effectstest.ConduitTable(t)
	effectstest.ConduitMana(t, g, me, "{C}{G}{R}")
	if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// #4250: the cast is made during the ability's resolution, paid then.
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, bear)
	if err := conduitCastFromGraveyard(g, me, bear); err != nil {
		t.Fatalf("cast from graveyard: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, bear)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); !conduitBanned(err) {
		t.Fatalf("a second spell this turn must be refused by the ban, got %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the spell cast through Conduit resolves")
	}
}

func TestConduitOfWorldsDeclinedOfferBansNothing(t *testing.T) {
	g, me, conduit, bear, _, bolt := effectstest.ConduitTable(t)
	effectstest.ConduitMana(t, g, me, "{C}{G}{R}")
	if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// #4250: the decline is the resolution_cast's empty answer, and the
	// card stays in the graveyard (CR 608.2g).
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("the declined card left the graveyard")
	}
	if err := conduitCastFromGraveyard(g, me, bear); err == nil {
		t.Fatal("the offer closed on the decline; the card cannot be cast afterwards")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("declining the cast binds nobody: %v", err)
	}
}

func TestConduitOfWorldsExpiredOfferRunsNoFollowUp(t *testing.T) {
	g, me, conduit, bear, _, _ := effectstest.ConduitTable(t)
	if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The offer is declined inside the resolution, and the turn moves
	// on: nothing follows.
	effectstest.DeclineResolutionCast(t, g, me.ID)
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
	g, me, conduit, bear, _, bolt := effectstest.ConduitTable(t)
	effectstest.ConduitMana(t, g, me, "{C}{G}{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := conduitCastFromGraveyard(g, me, bear); err == nil {
		t.Error("\"if you haven't cast a spell this turn\": no offer after a spell was cast")
	}
}

func TestConduitOfWorldsOfferEndsWhenAnotherSpellIsCastFirst(t *testing.T) {
	g, me, conduit, bear, _, bolt := effectstest.ConduitTable(t)
	effectstest.ConduitMana(t, g, me, "{C}{G}{R}")
	if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
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
	g, me, conduit, _, land, _ := effectstest.ConduitTable(t)
	spell := effectstest.PushGraveyardCardWithTypeLineID(me, "Shock", "Instant")
	theirs := effectstest.PushGraveyardCardWithTypeLineID(g.Seats[1], "Their Bear", "Creature — Bear")
	for _, id := range []uuid.UUID{land, spell, theirs} {
		if err := effectstest.ConduitActivate(t, g, me, conduit, id); err == nil {
			t.Errorf("%s is not a legal target", id)
		}
	}
}

func TestConduitOfWorldsPlaysLandsFromTheGraveyard(t *testing.T) {
	g, me, _, _, land, _ := effectstest.ConduitTable(t)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("the standing permission lets you play a land from the graveyard: %v", err)
	}
}

func TestConduitOfWorldsEnumeratorAgrees(t *testing.T) {
	g, me, conduit, bear, _, _ := effectstest.ConduitTable(t)
	effectstest.ConduitMana(t, g, me, "{C}{G}")
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
	if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
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
