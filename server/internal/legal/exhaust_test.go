package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exhaust_test.go — the enumerator half of #1181. An exhaust ability
// this object has already activated is not a move: the engine refuses
// it with ErrAbilityExhausted, so a bot seat must never be offered it
// (#544's rule, and the reason the enumerator and the engine read the
// same game.AbilityExhausted rather than two copies of the rule).

const (
	oracleProwcatcherSpecialist = "594132ac-32a7-41d5-b7f0-3692bbed7f6f"
	oracleGreenbeltGuardian     = "2d8aa053-289d-40d9-baa7-9bd1c5b8e957"
)

func TestASpentExhaustAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	goblin := battlefieldCard(g, active, game.Card{
		Name: "Prowcatcher Specialist", TypeLine: "Creature — Goblin Warrior",
		ManaCost: "{1}{R}", Power: 2, Toughness: 2, OracleID: oracleProwcatcherSpecialist,
	})
	basics(g, active, "Mountain", 4)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, goblin)
	if len(acts) != 1 {
		t.Fatalf("an unspent exhaust ability: want 1 activation offered, got %v", labels(acts))
	}
	// Every offer is one the dispatcher accepts.
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, goblin, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the exhaust ability: %v", err)
	}
	basics(g, active, "Mountain", 4)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), goblin); len(acts) != 0 {
		t.Errorf("a spent exhaust ability: want no activation offered, got %v — the engine "+
			"refuses it, so nothing may propose it", labels(acts))
	}
}

// TestSpendingAnExhaustAbilityLeavesTheOtherOnesEnumerated is the
// per-ABILITY half on the printed card that has both: Greenbelt
// Guardian's repeatable "{G}: target creature gains trample" is still
// a move after its exhaust ability is gone.
func TestSpendingAnExhaustAbilityLeavesTheOtherOnesEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	elf := battlefieldCard(g, active, game.Card{
		Name: "Greenbelt Guardian", TypeLine: "Creature — Elf Ranger",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, OracleID: oracleGreenbeltGuardian,
	})
	basics(g, active, "Forest", 6)
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), elf); len(acts) < 2 {
		t.Fatalf("both abilities should be offered, got %v", labels(acts))
	}
	// Index 1 is the exhaust ability; index 0 is the repeatable pump.
	if err := g.ActivateCatalogAbility(active.ID, elf, 1, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the exhaust ability: %v", err)
	}
	basics(g, active, "Forest", 6)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), elf)
	if len(acts) == 0 {
		t.Fatal("the repeatable ability stopped being a move when the exhaust ability was spent")
	}
	for _, m := range acts {
		if m.Label == "Exhaust — {3}{G}: Put three +1/+1 counters on this creature." {
			t.Errorf("the spent exhaust ability is still offered: %v", labels(acts))
		}
	}
	dispatchAll(t, g, active.ID, acts)
}

// basics seeds n untapped basics of one type for a seat — the
// coloured sibling of `mana`, which seeds Islands and so cannot pay a
// {R} or a {G}.
func basics(g *game.Game, p *game.Player, name string, n int) {
	for i := 0; i < n; i++ {
		battlefieldCard(g, p, basic(name, name))
	}
}
