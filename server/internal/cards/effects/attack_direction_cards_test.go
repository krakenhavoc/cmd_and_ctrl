package effects

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_direction_cards_test.go — ADR 0206 (#4197): Pramikon, Sky
// Rampart and Teyo, Geometric Tactician's −2, "each player may attack
// only the nearest opponent in the chosen direction and planeswalkers
// controlled by that opponent".

const pramikonOracle = "377d0fa7-1b3d-4b97-b066-dceeff3734d0"

// directionTargets is what the enumerator offers `id` to attack.
func directionTargets(t *testing.T, g *game.Game, id uuid.UUID) []uuid.UUID {
	t.Helper()
	var out []uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardPtrForEffect(id)
		if !ok {
			t.Fatalf("%s is not on the battlefield", id)
		}
		for _, r := range g.AttackTargetsForAttackerForEffect(c) {
			out = append(out, r.ID)
		}
	})
	return out
}

// oneCreaturePerSeat gives every seat a creature to ask about.
func oneCreaturePerSeat(g *game.Game) []uuid.UUID {
	out := make([]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = pushCatalogPermanent(g, p.ID, "Bear", "Creature — Bear", "", false)
	}
	return out
}

func TestPramikonAndTeyoAreFull(t *testing.T) {
	for _, oracle := range []string{pramikonOracle, guffcTeyoOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Pramikon chooses as it enters, and from then on every player, its
// controller included, may attack only the nearest opponent that way
// and their planeswalkers.
func TestPramikonLimitsEveryPlayerToTheNearestOpponentInTheChosenDirection(t *testing.T) {
	g, me, _ := spendTable(t)
	bears := oneCreaturePerSeat(g)
	seats := g.Seats
	n := len(seats)
	pramikon := castCatalogSpell(t, g, "Pramikon, Sky Rampart", "Legendary Creature — Wall", pramikonOracle, nil)
	passPriorityAroundTable(t, g)

	// Until the controller answers, nothing is restricted.
	if got := directionTargets(t, g, bears[1]); len(got) != n-1 {
		t.Fatalf("before the choice seat 1 may attack %v, want all %d opponents", got, n-1)
	}
	answerAnchorWord(t, g, me.ID, "Right")
	if got := g.ChosenOptionOf(pramikon); got != "Right" {
		t.Fatalf("Pramikon chose %q, want Right", got)
	}
	walker := pushCatalogPermanent(g, seats[0].ID, "Walker", "Legendary Planeswalker — Test", "", false)
	for i := range seats {
		right := seats[(i+n-1)%n].ID
		got := directionTargets(t, g, bears[i])
		want := []uuid.UUID{right}
		if right == seats[0].ID {
			want = append(want, walker)
		}
		if len(got) != len(want) {
			t.Errorf("seat %d may attack %v, want %v (the player before them)", i, got, want)
			continue
		}
		for _, w := range want {
			if !slices.Contains(got, w) {
				t.Errorf("seat %d may attack %v, missing %s", i, got, w)
			}
		}
	}

	// The verbs agree, and the refusal names Pramikon.
	advanceToDeclareAttackersOf(t, g, 0)
	err := g.Clone().DeclareAttacker(bears[0], seats[1].ID)
	var de *game.AttackDirectionError
	if !errors.As(err, &de) || !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("attack to the left: err = %v, want an AttackDirectionError", err)
	}
	if s := de.Sentence(); !strings.Contains(s, "to the right") || !strings.Contains(s, "(Pramikon, Sky Rampart)") {
		t.Errorf("sentence = %q", s)
	}
	if err := g.Clone().DeclareAttacker(bears[0], seats[n-1].ID); err != nil {
		t.Errorf("attack to the right: %v", err)
	}
}

// A Pramikon that never chose (a copy that entered without choosing)
// restricts nothing: the Pramikon ruling.
func TestPramikonWithNoChoiceRestrictsNothing(t *testing.T) {
	g, me, _ := spendTable(t)
	bears := oneCreaturePerSeat(g)
	pushCatalogPermanent(g, me.ID, "Pramikon, Sky Rampart", "Legendary Creature — Wall", pramikonOracle, false)
	if got := directionTargets(t, g, bears[1]); len(got) != len(g.Seats)-1 {
		t.Errorf("seat 1 may attack %v, want every opponent", got)
	}
}

// Teyo's −2: its controller chooses as it resolves, every player is held
// to that direction through the other players' turns, and the rule ends
// as Teyo's controller's next turn begins.
func TestTeyoMinusTwoLimitsAttacksUntilYourNextTurn(t *testing.T) {
	g, me, _ := spendTable(t)
	teyo := guffcWalker(t, g, me.ID, "Teyo, Geometric Tactician", "Legendary Planeswalker — Teyo", guffcTeyoOracle, 3)
	bears := oneCreaturePerSeat(g)
	seats := g.Seats
	n := len(seats)

	b16Activate(t, g, me.ID, teyo, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, me.ID, "Left")

	for i := range seats {
		left := seats[(i+1)%n].ID
		got := directionTargets(t, g, bears[i])
		want := []uuid.UUID{left}
		if left == me.ID {
			want = append(want, teyo)
		}
		if len(got) != len(want) {
			t.Errorf("seat %d may attack %v, want %v (the next player)", i, got, want)
		}
		for _, w := range want {
			if !slices.Contains(got, w) {
				t.Errorf("seat %d may attack %v, missing %s", i, got, w)
			}
		}
	}

	advanceToDeclareAttackersOf(t, g, 1)
	err := g.Clone().DeclareAttacker(bears[1], me.ID)
	var de *game.AttackDirectionError
	if !errors.As(err, &de) || !strings.Contains(de.Sentence(), "(Teyo, Geometric Tactician)") {
		t.Fatalf("seat 1 at Teyo's controller: err = %v, want a refusal naming Teyo", err)
	}
	advanceToDeclareAttackersOf(t, g, n-1)
	if got := directionTargets(t, g, bears[n-1]); len(got) != 2 {
		t.Errorf("on the last seat's turn it may attack %v, want Teyo's controller and Teyo", got)
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if got := directionTargets(t, g, bears[0]); len(got) != n-1 {
		t.Errorf("on Teyo's controller's next turn they may attack %v, want every opponent", got)
	}
}
