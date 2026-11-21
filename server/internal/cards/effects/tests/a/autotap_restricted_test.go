package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestAutoPayTapsCavernOfSoulsForACreatureOfTheChosenType(t *testing.T) {
	t.Run("an Elf off the Cavern alone can't be countered", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		cavern := effectstest.PushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", effectstest.CavernOfSoulsOracle, "Elf")

		elf := effectstest.CastFromHandForTest(t, g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", effectstest.AutoPay)
		if !effectstest.PermanentTapped(g, cavern) {
			t.Fatal("auto-pay did not tap the Cavern for the Elf")
		}
		if !g.StackMeta[elf].SpellCantBeCounteredByMana() {
			t.Fatalf("the Elf does not record the Cavern rider: %+v", g.StackMeta[elf].Paid.Mana)
		}
		if !effectstest.CounterByEffect(t, g, elf) {
			t.Error("the auto-paid Elf was countered")
		}
	})

	t.Run("the coloured half is preferred to the {C} half and to a Forest", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		forest := effectstest.RestrictedBasic(t, g, me, "Forest")
		cavern := effectstest.PushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", effectstest.CavernOfSoulsOracle, "Elf")

		elf := effectstest.CastFromHandForTest(t, g, me, "Elvish Visionary", "Creature — Elf Shaman", "{1}{G}", "", effectstest.AutoPay)
		if !effectstest.PermanentTapped(g, cavern) || !effectstest.PermanentTapped(g, forest) {
			t.Fatal("a {1}{G} Elf off a Forest and a Cavern did not tap both")
		}
		if !effectstest.CounterByEffect(t, g, elf) {
			t.Error("the Cavern paid with its {C} half: the Elf was countered")
		}
	})

	t.Run("a Goblin is paid by the Mountain and the Cavern is left", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		cavern := effectstest.PushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", effectstest.CavernOfSoulsOracle, "Elf")
		mountain := effectstest.RestrictedBasic(t, g, me, "Mountain")

		goblin := effectstest.CastFromHandForTest(t, g, me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", effectstest.AutoPay)
		if effectstest.PermanentTapped(g, cavern) || !effectstest.PermanentTapped(g, mountain) {
			t.Fatal("want the Mountain tapped and the Cavern untapped for a Goblin")
		}
		if effectstest.CounterByEffect(t, g, goblin) {
			t.Error("the Goblin could not be countered")
		}
	})

	t.Run("a Goblin off the Cavern alone is refused", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		cavern := effectstest.PushNamedTribePermanent(t, g, me.ID, "Cavern of Souls", "Land", effectstest.CavernOfSoulsOracle, "Elf")

		_, err := effectstest.CastFromHandErr(g, me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", effectstest.AutoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Fatalf("cast = %v, want insufficient mana", err)
		}
		if effectstest.PermanentTapped(g, cavern) {
			t.Error("the refused Goblin tapped the Cavern")
		}
	})
}

func TestAutoPayTapsDelightedHalflingForALegendarySpell(t *testing.T) {
	board := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		halfling := effectstest.SeedPermanentWithOracle(g, me.ID, "Delighted Halfling", "Creature — Halfling Citizen", effectstest.DelightedHalflingOracl)
		forest := effectstest.RestrictedBasic(t, g, me, "Forest")
		return g, me, halfling, forest
	}

	t.Run("a legendary spell is paid by the Halfling and can't be countered", func(t *testing.T) {
		g, me, halfling, forest := board(t)
		yeva := effectstest.CastFromHandForTest(t, g, me, "Yeva, Nature's Herald", "Legendary Creature — Elf Shaman", "{G}", "", effectstest.AutoPay)
		if !effectstest.PermanentTapped(g, halfling) || effectstest.PermanentTapped(g, forest) {
			t.Fatal("want the Halfling tapped and the Forest untapped for a legendary spell")
		}
		if !effectstest.CounterByEffect(t, g, yeva) {
			t.Error("the legendary spell was countered")
		}
	})

	t.Run("a nonlegendary spell is paid by the Forest and the Halfling is left", func(t *testing.T) {
		g, me, halfling, forest := board(t)
		bear := effectstest.CastFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{G}", "", effectstest.AutoPay)
		if effectstest.PermanentTapped(g, halfling) || !effectstest.PermanentTapped(g, forest) {
			t.Fatal("want the Forest tapped and the Halfling untapped for a nonlegendary spell")
		}
		if effectstest.CounterByEffect(t, g, bear) {
			t.Error("the nonlegendary spell could not be countered")
		}
	})

	t.Run("a nonlegendary {G}{G} cannot use the Halfling's colour", func(t *testing.T) {
		g, me, halfling, _ := board(t)
		_, err := effectstest.CastFromHandErr(g, me, "Leatherback Baloth", "Creature — Beast", "{G}{G}", "", effectstest.AutoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Fatalf("cast = %v, want insufficient mana", err)
		}
		if effectstest.PermanentTapped(g, halfling) {
			t.Error("the refused cast tapped the Halfling")
		}
	})
}
