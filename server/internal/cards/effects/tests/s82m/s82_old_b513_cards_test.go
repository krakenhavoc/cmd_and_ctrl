package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b513_cards_test.go — Shards of Alara, batch 513 (S82 older sets,
// #5588). Every test plays the real catalog entry through the engine.

const (
	oldB513KeeperOracle      = "a30241b4-5e92-44cd-b19b-7d6d59e95ada"
	oldB513InfiltratorOracle = "6ec8dc98-adc7-4fc3-94e5-31037235ceab"
)

func TestOldB513EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		"220bfa3b-0e17-4350-a967-cfcf05b754ef": "Woolly Thoctar",
		"2d425f5d-326f-46f8-ab84-852b01db4adb": "Cylian Elf",
		oldB513InfiltratorOracle:               "Jhessian Infiltrator",
		"94190aaf-db91-4a3c-a2a7-e90868e856ff": "Incurable Ogre",
		oldB513KeeperOracle:                    "Keeper of Progenitus",
		"acc07d0d-141b-4169-9643-6f20739e91be": "Jhessian Lookout",
		"ecc5467d-d327-4a9c-b935-09afbea111dc": "Dreg Reaver",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB513JhessianInfiltratorCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	inf := effectstest.B12Push(g, me.ID, "Jhessian Infiltrator", "Creature — Human Rogue", oldB513InfiltratorOracle, 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{inf, plain} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := effectstest.BlockRefusal(t, g, blocker, inf); got == "" {
		t.Error("the Infiltrator was blockable")
	}
	if got := effectstest.BlockRefusal(t, g, blocker, plain); got != "" {
		t.Errorf("an ordinary attacker was refused a block: %q", got)
	}
}

func TestOldB513KeeperDoublesMountainForestAndPlainsForEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Keeper of Progenitus", "Creature — Elf Druid", oldB513KeeperOracle, 1, 3)
	for _, tc := range []struct {
		name, typeLine, color string
		who                   int
		want                  int
	}{
		{"Forest", "Basic Land — Forest", "G", 0, 2},
		{"Mountain", "Basic Land — Mountain", "R", 1, 2},
		{"Plains", "Basic Land — Plains", "W", 1, 2},
		{"Island", "Basic Land — Island", "U", 0, 1},
		{"Swamp", "Basic Land — Swamp", "B", 1, 1},
	} {
		p := g.Seats[tc.who]
		before := len(effectstest.PoolColors(p))
		land := effectstest.B12Permanent(g, p.ID, tc.name, tc.typeLine)
		effectstest.TapForMana(t, g, p.ID, land)
		got := effectstest.PoolColors(p)
		if len(got)-before != tc.want {
			t.Errorf("%s: pool grew by %d, want %d (%v)", tc.name, len(got)-before, tc.want, got)
		}
		for _, c := range got[before:] {
			if c != tc.color {
				t.Errorf("%s: mana %q, want %q", tc.name, c, tc.color)
			}
		}
	}
	_ = them
}
