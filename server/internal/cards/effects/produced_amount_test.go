package effects

import (
	"runtime"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// produced_amount_test.go — #3042 from the card side: the count-driven
// helpers return a game.ManaAmount, the engine consumes the count, the
// ceiling still announces, and a replacement still edits the mana.

const (
	gaeasCradleOracleForAmount    = "7c427c3d-ecd8-45ef-bebd-8f10f4a311db"
	cityOfShadowsOracleForAmount  = "043192d8-6077-46c3-b43f-b7caf6762869"
	throneOfEldraineOracleForAmt  = "5b9a2b81-a645-43be-8001-03817ac210ca"
	priestOfTitaniaOracleForAmont = "3a198a16-17b9-481e-b516-5bc945c7e247"
)

func amountOf(t *testing.T, g *game.Game, oracle string, source, controller uuid.UUID) game.ManaAmount {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok || len(spec.ManaAbilities) == 0 || spec.ManaAbilities[0].ProducedAmount == nil {
		t.Fatalf("%s has no ProducedAmount mana ability", oracle)
	}
	var a game.ManaAmount
	g.ReadSnapshot(func() { a = spec.ManaAbilities[0].ProducedAmount(g, controller, source) })
	return a
}

func TestProducedAmountHelpersReturnTheCount(t *testing.T) {
	t.Run("per permanent", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		cradle := pushCatalogPermanent(g, me.ID, "Gaea's Cradle", "Legendary Land", gaeasCradleOracleForAmount, false)
		for i := 0; i < 3; i++ {
			pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
				Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
		}
		a := amountOf(t, g, gaeasCradleOracleForAmount, cradle, me.ID)
		if a.N != 3 || len(a.Options) != 1 || a.Options[0] != "G" || a.OneColor {
			t.Errorf("Cradle = %+v, want three {G}", a)
		}
	})

	t.Run("per counter on this", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushCatalogPermanent(g, me.ID, "City of Shadows", "Land", cityOfShadowsOracleForAmount, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterStorage, 4) })
		a := amountOf(t, g, cityOfShadowsOracleForAmount, id, me.ID)
		if a.N != 4 || len(a.Options) != 1 || a.Options[0] != "C" {
			t.Errorf("City of Shadows = %+v, want four {C}", a)
		}
	})

	t.Run("chosen colour amount", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushCatalogPermanent(g, me.ID, "Throne of Eldraine", "Legendary Artifact", throneOfEldraineOracleForAmt, false)
		if a := amountOf(t, g, throneOfEldraineOracleForAmt, id, me.ID); a.N != 0 {
			t.Errorf("before a colour is chosen the Throne = %+v, want nothing", a)
		}
	})

	t.Run("any one colour", func(t *testing.T) {
		fn := ProducedOneColor(func(*game.Game, uuid.UUID, uuid.UUID) int { return 6 })
		a := fn(nil, uuid.Nil, uuid.Nil)
		if a.N != 6 || !a.OneColor || len(a.Options) != 5 {
			t.Errorf("one colour = %+v", a)
		}
	})

	t.Run("any combination", func(t *testing.T) {
		fn := ProducedAnyCombinationOfColors(func(*game.Game, uuid.UUID, uuid.UUID) int { return 6 })
		a := fn(nil, uuid.Nil, uuid.Nil)
		if a.N != 6 || a.OneColor || len(a.Options) != 5 {
			t.Errorf("any combination = %+v", a)
		}
	})
}

// Marwyn is the hand-written row: power is the count, and the count is
// what the engine reads.
func TestMarwynIsDrivenByPowerAsACount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marwyn := pushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", b10MarwynOracle, false)
	ouroboroidSetCounters(g, marwyn, 4) // power 1 + 4
	a := amountOf(t, g, b10MarwynOracle, marwyn, me.ID)
	if a.N != 5 || a.Options[0] != "G" {
		t.Errorf("Marwyn = %+v, want five {G}", a)
	}
	tapForMana(t, g, me.ID, marwyn)
	if got := len(me.ManaPool); got != 5 {
		t.Errorf("pool holds %d, want 5", got)
	}
}

// A replacement that edits the produced mana between the ability and the
// pool still applies to a count-driven ability: Mana Reflection doubles
// Marwyn's three.
func TestManaReflectionStillDoublesACountDrivenAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	marwyn := pushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", b10MarwynOracle, false)
	ouroboroidSetCounters(g, marwyn, 2) // power 3
	tapForMana(t, g, me.ID, marwyn)
	if got := len(me.ManaPool); got != 6 {
		t.Errorf("Mana-Reflected Marwyn made %d mana, want 6", got)
	}
}

// At the ceiling the count is cut and announced before any token exists,
// and the work done is the ceiling's, not the count's.
func TestCountDrivenCeilingAnnouncesAndStaysSmall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marwyn := pushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", b10MarwynOracle, false)
	ouroboroidSetCounters(g, marwyn, game.MaxCount)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	tapForMana(t, g, me.ID, marwyn)
	runtime.ReadMemStats(&after)

	if got := len(me.ManaPool); got != game.MaxProducedMana {
		t.Errorf("pool holds %d, want %d", got, game.MaxProducedMana)
	}
	if n := countCappedEvents(g, "mana"); n != 1 {
		t.Errorf("want one mana count_capped announcement, got %d", n)
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 8<<20 {
		t.Errorf("tapping a ceiling Marwyn allocated %d bytes", grew)
	}
}

// Priest of Titania counts every Elf on the battlefield, anyone's.
func TestPriestOfTitaniaCountsElves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	priest := pushCatalogPermanent(g, me.ID, "Priest of Titania", "Creature — Elf Druid", priestOfTitaniaOracleForAmont, false)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf Warrior",
		Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if a := amountOf(t, g, priestOfTitaniaOracleForAmont, priest, me.ID); a.N != 2 {
		t.Errorf("Priest = %+v, want two (every Elf on the battlefield)", a)
	}
}
