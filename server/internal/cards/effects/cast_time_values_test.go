package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_time_values_test.go — the engine half of #3781: a value a spell
// fixes as it is cast (CR 601.2i) is stamped on the stack item, survives
// the undo clone and a JSON snapshot round trip, rides a spell copy
// (CR 707.10), and is never stamped on a spell that declares none.
// Lifestream's Blessing is the proof card; its behaviour tests are in
// cloud_deck_b_test.go.

// castLifestreams casts Lifestream's Blessing from hand with a 4/4 and a
// 1/1 on the board and returns the spell's id.
func castLifestreams(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	b12Creature(g, me.ID, "Big", "Creature — Test", 4, 4)
	b12Creature(g, me.ID, "Small", "Creature — Test", 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(me, "Lifestream's Blessing", "Instant", lifestreamsBlessingOrcl)
	me.ManaPool.AddMana(append(cloudDeckMana("C", 4), cloudDeckMana("G", 2)...)...)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return card
}

func castValueOf(g *game.Game, id uuid.UUID, key string) (v int, ok bool) {
	g.ReadSnapshot(func() {
		if item := g.StackItemForEffect(id); item != nil {
			v, ok = item.CastValue(key)
		}
	})
	return v, ok
}

func TestCastTimeValueSurvivesUndoCloneAndSnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := castLifestreams(t, g, me)

	if v, ok := castValueOf(g, spell, lifestreamsBlessingX); !ok || v != 4 {
		t.Fatalf("live item: %d, %v; want 4", v, ok)
	}

	undo := g.Clone()
	if v, ok := castValueOf(undo, spell, lifestreamsBlessingX); !ok || v != 4 {
		t.Errorf("clone: %d, %v; want 4", v, ok)
	}
	// The clone owns its own map: changing the live one leaves it alone.
	g.WithWriteLock(func() { g.StackMeta[spell].CastValues[lifestreamsBlessingX] = 99 })
	if v, _ := castValueOf(undo, spell, lifestreamsBlessingX); v != 4 {
		t.Errorf("the clone shares the live map: %d", v)
	}
	g.RestoreFrom(undo)
	if v, _ := castValueOf(g, spell, lifestreamsBlessingX); v != 4 {
		t.Errorf("after undo: %d, want 4", v)
	}

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if v, ok := castValueOf(restored, spell, lifestreamsBlessingX); !ok || v != 4 {
		t.Fatalf("restored: %d, %v; want 4", v, ok)
	}
	// And the restored game resolves with the recorded value even after
	// the board changed: the Big creature is gone, the draw is still four.
	restoredMe := restored.Seats[restored.Turn.ActiveSeat]
	hand := restoredMe.Hand.Size()
	for _, c := range restored.BattlefieldCardsForEffect() {
		if c.Name == "Big" {
			restored.WithWriteLock(func() { _ = restored.DestroyPermanentForEffect(c.InstanceID) })
		}
	}
	passPriorityAroundTable(t, restored)
	if got := restoredMe.Hand.Size() - hand; got != 4 {
		t.Errorf("restored game drew %d, want 4 (the value recorded as it was cast)", got)
	}
}

func TestCastTimeValueRidesASpellCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := castLifestreams(t, g, me)
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	copies := 0
	g.ReadSnapshot(func() {
		for id, item := range g.StackMeta {
			if id == spell || item == nil || item.Kind != game.StackItemSpell {
				continue
			}
			copies++
			if v, ok := item.CastValue(lifestreamsBlessingX); !ok || v != 4 {
				t.Errorf("the copy's cast-time value = %d, %v; want the original's 4", v, ok)
			}
			// A separate map: the copy is its own object.
			if item.CastValues != nil {
				item.CastValues[lifestreamsBlessingX] = 7
			}
		}
		if v, _ := g.StackMeta[spell].CastValue(lifestreamsBlessingX); v != 4 {
			t.Errorf("the copy aliases the original's map: %d", v)
		}
	})
	if copies != 1 {
		t.Fatalf("found %d copies, want 1", copies)
	}
}

func TestASpellThatDeclaresNoCastTimeValueRecordsNone(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID}})
	g.ReadSnapshot(func() {
		if item := g.StackItemForEffect(id); item == nil || item.CastValues != nil {
			t.Errorf("a spell with no declaration carries %+v", item)
		}
	})
}

func TestRegisterRefusesAMalformedCastTimeValue(t *testing.T) {
	val := func(*game.Game, *game.StackItem) int { return 0 }
	for name, spec := range map[string]Spec{
		"blank key": {Name: "X", CastTimeValues: []game.CastTimeValue{{Value: val}}},
		"nil value": {Name: "X", CastTimeValues: []game.CastTimeValue{{Key: "k"}}},
		"duplicate": {Name: "X", CastTimeValues: []game.CastTimeValue{{Key: "k", Value: val}, {Key: "k", Value: val}}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("checkCastTimeValues accepted a malformed declaration")
				}
			}()
			checkCastTimeValues(spec)
		})
	}
}
