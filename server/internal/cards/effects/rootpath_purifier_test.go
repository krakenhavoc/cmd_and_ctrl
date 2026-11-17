package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const rootpathPurifierOracle = "442d06e6-8cf6-4980-b2bf-73cfb4d3a468"

func TestRootpathPurifierIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(rootpathPurifierOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// libraryIsBasic reads whether the library card is a basic land after
// the layers (the way a tutor's filter reads it).
func libraryIsBasic(g *game.Game, p *game.Player, id uuid.UUID) bool {
	var basic bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range p.Library.Cards {
			if c.InstanceID == id {
				basic = IsBasicLand(c)
			}
		}
	})
	return basic
}

func battlefieldIsBasic(g *game.Game, id uuid.UUID) bool {
	var basic bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				basic = IsBasicLand(c)
			}
		}
	})
	return basic
}

// Lands you control on the battlefield and land cards in your library are
// basic; an opponent's lands, your nonland library cards and the effect
// itself end when the Purifier leaves.
func TestRootpathPurifierMakesYourLandsAndLibraryLandsBasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myLib := seedLibraryTop(me, "Cave", "Land", "")
	myNonland := seedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	oppLib := seedLibraryTop(opp, "Their Cave", "Land", "")
	mineLand := pushCatalogPermanent(g, me.ID, "My Cave", "Land", "", false)
	theirLand := pushCatalogPermanent(g, opp.ID, "Their Cave", "Land", "", false)
	if battlefieldIsBasic(g, mineLand) || libraryIsBasic(g, me, myLib) {
		t.Fatal("setup: a nonbasic land read as basic before the Purifier")
	}
	purifier := pushCatalogPermanent(g, me.ID, "Rootpath Purifier", "Creature — Elf Druid", rootpathPurifierOracle, false)
	if !battlefieldIsBasic(g, mineLand) {
		t.Error("a land you control is not basic")
	}
	if !libraryIsBasic(g, me, myLib) {
		t.Error("a land card in your library is not basic")
	}
	if libraryIsBasic(g, me, myNonland) {
		t.Error("a nonland library card is basic")
	}
	if battlefieldIsBasic(g, theirLand) || libraryIsBasic(g, opp, oppLib) {
		t.Error("an opponent's land is basic")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(purifier); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if battlefieldIsBasic(g, mineLand) || libraryIsBasic(g, me, myLib) {
		t.Error("the lands stayed basic after the Purifier left")
	}
}
