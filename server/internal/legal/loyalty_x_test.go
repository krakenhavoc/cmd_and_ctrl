package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// loyalty_x_test.go — #1944: a −X loyalty cost is offered at every X
// from the floor to the loyalty there (CR 606.6), one move per X, each
// naming its own loyalty price, and every one is a move the engine
// accepts (#544).

func minusXWalker(loyalty int, targets *game.TargetSpec) game.Card {
	zero := 0
	return game.Card{
		Name:     "X Walker",
		TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: loyalty},
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:   "−X: do nothing",
			Cost:    game.AbilityCost{Loyalty: &zero, LoyaltyX: true},
			Targets: targets,
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

func TestLoyaltyXIsOfferedAtEveryXTheLoyaltyPays(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	src := battlefieldCard(g, active, minusXWalker(4, nil))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, src)
	if len(acts) != 5 {
		t.Fatalf("4 loyalty: want X = 0..4, got %v", labels(acts))
	}
	seen := map[int]bool{}
	for _, m := range acts {
		x := xValueOf(t, m)
		seen[x] = true
		loyalty := 0
		if m.Cost != nil {
			loyalty = m.Cost.Loyalty
		}
		if loyalty != -x {
			t.Errorf("X=%d: cost loyalty = %d, want %d", x, loyalty, -x)
		}
	}
	for x := 0; x <= 4; x++ {
		if !seen[x] {
			t.Errorf("X=%d not offered", x)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// TestLoyaltyXReachesEveryTargetAtEveryX: the ladder does not spend the
// per-source budget one X at a time, so the second target is offered at
// every X too.
func TestLoyaltyXReachesEveryTargetAtEveryX(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	creatureSpec := costSpec("target creature", func(c game.Card) bool { return c.IsCreature() })
	src := battlefieldCard(g, active, minusXWalker(3, creatureSpec))
	battlefieldCard(g, other, creature("Bear", "{1}{G}", 2, 2))
	battlefieldCard(g, other, creature("Ogre", "{2}{R}", 3, 3))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, src)
	if len(acts) != 8 {
		t.Fatalf("two targets, X = 0..3: want 8 moves, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}
