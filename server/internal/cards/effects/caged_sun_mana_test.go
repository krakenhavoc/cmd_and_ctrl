package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// caged_sun_mana_test.go — #1600: the mana-doubling half of Caged
// Sun's printed text, on top of the anthem half already covered by
// TestB295CagedSunPumpsCreaturesOfTheChosenColor in
// slice_295b_mana_rocks_test.go.

const cagedSunTaigaOracle = "22e3cf1d-3559-4ce1-954c-8dc815342979" // Taiga: {T}: Add {R} or {G}.

// A land of the chosen colour gives one extra mana of that colour.
func TestCagedSunAddsOneManaOfTheChosenColorFromALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushChosenColorPermanent(t, g, me, "Caged Sun", "Artifact", cagedSunOracle, "R")
	mountain := b31Push(g, me, "Mountain", "Basic Land — Mountain", "", "", 0, 0)

	tapForMana(t, g, me, mountain)
	if got := sortedPool(g.Seats[0]); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool = %v, want two red — the Mountain's and Caged Sun's additional one", got)
	}
}

// A land of another colour gives no bonus.
func TestCagedSunAddsNothingFromALandOfAnotherColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushChosenColorPermanent(t, g, me, "Caged Sun", "Artifact", cagedSunOracle, "R")
	forest := b31Push(g, me, "Forest", "Basic Land — Forest", "", "", 0, 0)

	tapForMana(t, g, me, forest)
	if got := sortedPool(g.Seats[0]); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want just the Forest's {G} — green isn't the chosen colour", got)
	}
}

// An opponent's land of the chosen colour gives that opponent no
// bonus: the clause reads "you add", and "you" is Caged Sun's
// controller.
func TestCagedSunIgnoresAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	them := g.Seats[1].ID
	pushChosenColorPermanent(t, g, me, "Caged Sun", "Artifact", cagedSunOracle, "R")
	theirMountain := b31Push(g, them, "Mountain", "Basic Land — Mountain", "", "", 0, 0)

	tapForMana(t, g, them, theirMountain)
	if got := sortedPool(g.Seats[1]); len(got) != 1 || got[0] != "R" {
		t.Errorf("their pool = %v, want just their Mountain's {R} — Caged Sun is mine, not theirs", got)
	}
}

// A dual land tapped for the chosen colour gets the bonus; the same
// dual tapped for its other colour does not.
func TestCagedSunDualLandChosenColorVersusOtherColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushChosenColorPermanent(t, g, me.ID, "Caged Sun", "Artifact", cagedSunOracle, "R")
	taiga := seedPermanentWithOracle(g, me.ID, "Taiga", "Land — Mountain Forest", cagedSunTaigaOracle)

	tapForMana(t, g, me.ID, taiga)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for Taiga's {R} or {G}")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice(R): %v", err)
	}
	if got := sortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool = %v, want two red — Taiga's chosen {R} plus Caged Sun's additional one", got)
	}

	untap(g, taiga)
	g.WithWriteLock(func() { me.ManaPool = nil })

	tapForMana(t, g, me.ID, taiga)
	pick = pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for Taiga's second tap")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice(G): %v", err)
	}
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want just {G} — green isn't Caged Sun's chosen colour", got)
	}
}

// A land that makes two mana of the chosen colour in one activation
// (Phyrexian Tower's sacrifice half) gets only one extra, not one per
// token produced.
func TestCagedSunOnlyAddsOneWhenALandMakesTwoOfTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushChosenColorPermanent(t, g, me.ID, "Caged Sun", "Artifact", cagedSunOracle, "B")
	tower := seedPermanentWithOracle(g, me.ID, "Phyrexian Tower", "Legendary Land", phyrexianTowerOracle)
	bear := seedCreature(g, "Bear", me.ID)

	if err := g.ActivateManaAbility(me.ID, tower, 1, game.ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := sortedPool(me); len(got) != 3 || got[0] != "B" || got[1] != "B" || got[2] != "B" {
		t.Errorf("pool = %v, want three black — the Tower's two plus Caged Sun's one additional", got)
	}
}

// Before a colour is chosen (the ETB prompt hasn't been answered),
// Caged Sun adds nothing — never "any colour".
func TestCagedSunAddsNothingBeforeAColorIsChosen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Caged Sun", TypeLine: "Artifact", OracleID: cagedSunOracle,
		Owner: me, Controller: me,
	})
	mountain := b31Push(g, me, "Mountain", "Basic Land — Mountain", "", "", 0, 0)

	tapForMana(t, g, me, mountain)
	if got := sortedPool(g.Seats[0]); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool = %v, want just {R} — no colour has been chosen yet", got)
	}
}
