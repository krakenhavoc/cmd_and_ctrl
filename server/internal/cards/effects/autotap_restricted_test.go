package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_restricted_test.go — #3874, ADR 0040 amendment 2026-10-16:
// auto-pay plans Cavern of Souls' and Delighted Halfling's coloured
// halves for a spell their restriction admits, and the spell can't be
// countered; for any other spell they are left alone. The engine half
// (the candidate rule, its ranking, the payability bound) is
// game/autotap_restricted_test.go; the enumerator's half is
// legal/restricted_autopay_test.go.

var autoPay = game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"}

// restrictedBasic puts a basic land on the battlefield.
func restrictedBasic(t *testing.T, g *game.Game, me *game.Player, name string) uuid.UUID {
	t.Helper()
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{Name: name, TypeLine: "Basic Land — " + name}, 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("spawn %s: %v", name, err)
	}
	return ids[0]
}

func TestAutoPayTapsCavernOfSoulsForACreatureOfTheChosenType(t *testing.T) {
	t.Run("an Elf off the Cavern alone can't be countered", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")

		elf := castFromHandForTest(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", autoPay)
		if !permanentTapped(g, cavern) {
			t.Fatal("auto-pay did not tap the Cavern for the Elf")
		}
		if !g.StackMeta[elf].SpellCantBeCounteredByMana() {
			t.Fatalf("the Elf does not record the Cavern rider: %+v", g.StackMeta[elf].Paid.Mana)
		}
		if !counterByEffect(t, g, elf) {
			t.Error("the auto-paid Elf was countered")
		}
	})

	t.Run("the coloured half is preferred to the {C} half and to a Forest", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		forest := restrictedBasic(t, g, me, "Forest")
		cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")

		elf := castFromHandForTest(t, g, me, "Elvish Visionary", "Creature — Elf Shaman", "{1}{G}", "", autoPay)
		if !permanentTapped(g, cavern) || !permanentTapped(g, forest) {
			t.Fatal("a {1}{G} Elf off a Forest and a Cavern did not tap both")
		}
		if !counterByEffect(t, g, elf) {
			t.Error("the Cavern paid with its {C} half: the Elf was countered")
		}
	})

	t.Run("a Goblin is paid by the Mountain and the Cavern is left", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")
		mountain := restrictedBasic(t, g, me, "Mountain")

		goblin := castFromHandForTest(t, g, me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", autoPay)
		if permanentTapped(g, cavern) || !permanentTapped(g, mountain) {
			t.Fatal("want the Mountain tapped and the Cavern untapped for a Goblin")
		}
		if counterByEffect(t, g, goblin) {
			t.Error("the Goblin could not be countered")
		}
	})

	t.Run("a Goblin off the Cavern alone is refused", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		cavern := pushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", cavernOfSoulsOracle, "Elf")

		_, err := castFromHandErr(g, me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", autoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Fatalf("cast = %v, want insufficient mana", err)
		}
		if permanentTapped(g, cavern) {
			t.Error("the refused Goblin tapped the Cavern")
		}
	})
}

func TestAutoPayTapsDelightedHalflingForALegendarySpell(t *testing.T) {
	board := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		halfling := seedPermanentWithOracle(g, me.ID, "Delighted Halfling", "Creature — Halfling Citizen", delightedHalflingOracl)
		forest := restrictedBasic(t, g, me, "Forest")
		return g, me, halfling, forest
	}

	t.Run("a legendary spell is paid by the Halfling and can't be countered", func(t *testing.T) {
		g, me, halfling, forest := board(t)
		yeva := castFromHandForTest(t, g, me, "Yeva, Nature's Herald", "Legendary Creature — Elf Shaman", "{G}", "", autoPay)
		if !permanentTapped(g, halfling) || permanentTapped(g, forest) {
			t.Fatal("want the Halfling tapped and the Forest untapped for a legendary spell")
		}
		if !counterByEffect(t, g, yeva) {
			t.Error("the legendary spell was countered")
		}
	})

	t.Run("a nonlegendary spell is paid by the Forest and the Halfling is left", func(t *testing.T) {
		g, me, halfling, forest := board(t)
		bear := castFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{G}", "", autoPay)
		if permanentTapped(g, halfling) || !permanentTapped(g, forest) {
			t.Fatal("want the Forest tapped and the Halfling untapped for a nonlegendary spell")
		}
		if counterByEffect(t, g, bear) {
			t.Error("the nonlegendary spell could not be countered")
		}
	})

	t.Run("a nonlegendary {G}{G} cannot use the Halfling's colour", func(t *testing.T) {
		g, me, halfling, _ := board(t)
		_, err := castFromHandErr(g, me, "Leatherback Baloth", "Creature — Beast", "{G}{G}", "", autoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Fatalf("cast = %v, want insufficient mana", err)
		}
		if permanentTapped(g, halfling) {
			t.Error("the refused cast tapped the Halfling")
		}
	})
}
