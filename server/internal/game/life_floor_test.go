package game

import "testing"

// life_floor_test.go — #5770: the DERIVED half of the life floor, a
// permanent's printed "damage that would reduce your life total to less
// than 1 reduces it to 1 instead" (Sustaining Spirit), with the catalog
// hook stubbed.

func withCatalogLifeFloor(t *testing.T, fn func(oracleID string) bool) {
	t.Helper()
	prev := CatalogPlayerLifeFloorOne
	CatalogPlayerLifeFloorOne = fn
	t.Cleanup(func() { CatalogPlayerLifeFloorOne = prev })
}

// Combat damage from a commander: the life it costs stops at 1, and the
// CR 903.10a tally and lifelink (CR 702.15b) still see all of it.
func TestAPrintedLifeFloorKeepsCombatAndCommanderDamageWhole(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const spirit = "test-life-floor-spirit"
	withCatalogLifeFloor(t, func(id string) bool { return id == spirit })
	pushLeyline(t, g, me, spirit)

	cmdr := pushColouredCreature(g, opp, "General", []string{"W"}, "lifelink")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	me.Life = 4
	oppLife := opp.Life
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(me.ID, cmdr, 7, "") })
	if me.Life != 1 {
		t.Errorf("life after 7 combat damage from 4 under the floor = %d, want 1", me.Life)
	}
	if got := me.CommanderDamage[cmdr]; got != 7 {
		t.Errorf("commander damage tally = %d, want 7 — CR 903.10a counts damage dealt", got)
	}
	if opp.Life != oppLife+7 {
		t.Errorf("lifelink gained %d, want the full 7", opp.Life-oppLife)
	}
}

// Two copies compose, one leaving keeps the other's floor, and the floor
// ends with the last of them. Nothing is written to the player.
func TestAPrintedLifeFloorComposesAndEndsWithItsPermanent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const spirit = "test-life-floor-spirit-two"
	withCatalogLifeFloor(t, func(id string) bool { return id == spirit })
	first := pushLeyline(t, g, me, spirit)
	second := pushLeyline(t, g, me, spirit)

	floored := func() bool {
		var out bool
		g.ReadSnapshot(func() { out = g.playerLifeFloorLocked(me) })
		return out
	}
	if !floored() {
		t.Fatal("two printed floors and no floor")
	}
	var oppFloored bool
	g.ReadSnapshot(func() { oppFloored = g.playerLifeFloorLocked(opp) })
	if oppFloored {
		t.Error("the floor protected a seat that does not control the permanent")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, first); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if !floored() {
		t.Error("one copy leaving removed the floor the other still gives")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, second); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if floored() {
		t.Error("the floor outlived the permanents that printed it")
	}
	if n := len(me.Statics); n != 0 {
		t.Errorf("the derived floor wrote %d entries to Player.Statics; it must write none", n)
	}
}

// The floor is a static ability of the permanent, read through
// CatalogAbilityKey: one that has lost all its abilities (CR 613.1f) gives
// none.
func TestAPrintedLifeFloorEndsWhenThePermanentLosesItsAbilities(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	const spirit = "test-life-floor-spirit-blanked"
	withCatalogLifeFloor(t, func(id string) bool { return id == spirit })
	id := pushLeyline(t, g, me, spirit)

	registerScopedEffectForTest(t, g, id, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	var floored bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		floored = g.playerLifeFloorLocked(me)
	})
	if floored {
		t.Error("a permanent that lost all its abilities still gives the floor")
	}
}
