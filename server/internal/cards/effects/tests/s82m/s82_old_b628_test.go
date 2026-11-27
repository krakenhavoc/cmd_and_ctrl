package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b628_test.go — Judgment cards, S82 OLD batch 628 (#5611), played
// through the real catalog entries.

const (
	o628FuneralPyre = "02af53c1-83e6-447f-908b-2757c8aa445c"
	o628Riftstone   = "8d7e05ba-5406-4d5e-bb8f-a4a6f3b0eaa7"
	o628Orgg        = "cbbcd3a3-6ddd-4511-a352-3055d0102e7c"
)

func TestS82OldB628CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o628FuneralPyre: "Funeral Pyre", o628Riftstone: "Riftstone Portal", o628Orgg: "Soulgorger Orgg",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB628FuneralPyreGivesTheCardsOwnerASpirit(t *testing.T) {
	for _, theirs := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		owner := me
		if theirs {
			owner = opp
		}
		card := effectstest.PushGraveyardCardWithTypeLineID(owner, "Dead Card", "Sorcery")
		effectstest.CastCatalogSpell(t, g, "Funeral Pyre", "Instant", o628FuneralPyre, []game.TargetRef{{Kind: game.TargetCard, ID: card}})
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(card) {
			t.Fatalf("theirs=%v: the card was not exiled", theirs)
		}
		if got := effectstest.CountOnBattlefield(g, "Spirit", owner.ID); got != 1 {
			t.Errorf("theirs=%v: the card's owner has %d Spirit tokens, want 1", theirs, got)
		}
		other := me
		if !theirs {
			other = opp
		}
		if got := effectstest.CountOnBattlefield(g, "Spirit", other.ID); got != 0 {
			t.Errorf("theirs=%v: the other player has %d Spirit tokens, want 0", theirs, got)
		}
	}
}

func TestS82OldB628RiftstonePortalTapsForColorlessAndGrantsGreenOrWhiteFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	portal := effectstest.B12Push(g, me.ID, "Riftstone Portal", "Land", o628Riftstone, 0, 0)
	forest := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := effectstest.GaPut(g, opp.ID, "Their Land", "Basic Land — Island", "")
	if n := effectstest.GrantedManaRowsOf(g, forest); n != 0 {
		t.Fatalf("with the Portal on the battlefield the Forest has %d granted rows, want 0", n)
	}
	if err := g.ActivateManaAbility(me.ID, portal, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Portal: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.KillCreature(t, g, me.ID, portal)
	if !me.Graveyard.Contains(portal) {
		t.Fatal("the Portal is not in its owner's graveyard")
	}
	if n := effectstest.GrantedManaRowsOf(g, forest); n != 1 {
		t.Fatalf("with the Portal in the graveyard the Forest has %d granted rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's land has %d granted rows, want 0", n)
	}
	for _, color := range []string{"G", "W"} {
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == forest {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
		})
		if err := effectstest.TapGrantedMana(t, g, me.ID, forest, color); err != nil {
			t.Fatalf("tap the Forest for %s: %v", color, err)
		}
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != color {
			t.Errorf("pool %v, want [%s]", got, color)
		}
	}
}

func TestS82OldB628SoulgorgerOrggLosesAllButOneLifeAndGivesItBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	start := me.Life
	orgg := effectstest.CastCatalogSpell(t, g, "Soulgorger Orgg", "Creature — Nightmare Orgg", o628Orgg, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 1 {
		t.Fatalf("life after entering = %d, want 1", me.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(orgg) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start {
		t.Errorf("life after leaving = %d, want %d", me.Life, start)
	}
}

func TestS82OldB628SoulgorgerOrggAtOneLifeLosesAndGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Life = 1 })
	orgg := effectstest.CastCatalogSpell(t, g, "Soulgorger Orgg", "Creature — Nightmare Orgg", o628Orgg, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(orgg) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 1 {
		t.Errorf("life = %d, want 1", me.Life)
	}
}

func TestS82OldB628SoulgorgerOrggGainsOnlyWhatItsOwnEntryCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	start := me.Life
	orgg := effectstest.CastCatalogSpell(t, g, "Soulgorger Orgg", "Creature — Nightmare Orgg", o628Orgg, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Gaining life in the meantime does not change what the entry cost.
	g.WithWriteLock(func() { me.Life += 5 })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(orgg) })
	effectstest.PassPriorityAroundTable(t, g)
	if want := start + 5; me.Life != want {
		t.Errorf("life = %d, want %d", me.Life, want)
	}
}
