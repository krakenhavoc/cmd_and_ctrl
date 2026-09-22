package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// resolution_pick_test.go — #1214's three kinds through the enumerator.
//
// The enumeration test in choice_gate_test.go fails on a kind with no
// CASE. This is the other half: the case offers answers the engine
// accepts, the seat owing one is offered nothing else, and the pool is
// walked in the order #1014's hooks ask for — which on a real board is
// the difference between the move list containing the permanent that
// matters and not containing it at all.

// pickBoard puts n creatures under a seat's control, the i-th with
// power i+1, and returns their IDs plus the biggest.
func pickBoard(g *game.Game, p *game.Player, n int) (all []uuid.UUID, biggest uuid.UUID) {
	for i := 0; i < n; i++ {
		id := battlefieldCard(g, p, creature("Pick Bear", "{G}", i+1, i+1))
		all = append(all, id)
		biggest = id
	}
	return all, biggest
}

// queueOwnPick asks `seat` to choose one of its own permanents.
func queueOwnPick(t *testing.T, g *game.Game, seat uuid.UUID, lo, hi int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenForEffect(game.PermanentPickPrompt{
			Chooser:    seat,
			Question:   "test — choose your own",
			Of:         []uuid.UUID{seat},
			Candidates: pickCandidates(lo, hi),
		}, func(*game.Game, game.PromptedPicks) error { return nil }); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
				id = c.ID
			}
		}
	})
	return id
}

// queueTheirPick asks `seat` to choose among `of`'s permanents.
func queueTheirPick(t *testing.T, g *game.Game, seat, of uuid.UUID, lo, hi int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenForEffect(game.PermanentPickPrompt{
			Chooser:    seat,
			Question:   "test — choose theirs",
			Of:         []uuid.UUID{of},
			Candidates: pickCandidates(lo, hi),
		}, func(*game.Game, game.PromptedPicks) error { return nil }); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTheirPermanents {
				id = c.ID
			}
		}
	})
	return id
}

func pickCandidates(lo, hi int) func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	return func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
		var out []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == of {
				out = append(out, c.InstanceID)
			}
		}
		return out, lo, hi
	}
}

// powerScorer is a stand-in for a policy's scorer: power, read off the
// authoritative board, because these tests are about the HOOK rather
// than about the heuristic.
func powerScorer(g *game.Game) func(legal.TargetCandidate) float64 {
	power := map[uuid.UUID]float64{}
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			power[g.Battlefield.Cards[i].InstanceID] = float64(g.Battlefield.Cards[i].Power)
		}
	})
	return func(c legal.TargetCandidate) float64 { return power[c.ID] }
}

// TestResolutionPicksAreEnumeratedAndAnswerable — the #499 / #618
// contract for all three kinds at once: the seat owing one is offered
// that prompt's answers and nothing else, and every one of them is an
// answer the engine accepts.
func TestResolutionPicksAreEnumeratedAndAnswerable(t *testing.T) {
	t.Run("own_permanents", func(t *testing.T) {
		g := newTable(t)
		seat := g.Seats[0]
		pickBoard(g, seat, 3)
		queueOwnPick(t, g, seat.ID, 1, 1)
		moves := legal.EnumerateFor(g, seat.ID)
		assertOnlyChoiceMoves(t, g, seat.ID, moves)
	})
	t.Run("their_permanents", func(t *testing.T) {
		g := newTable(t)
		seat, other := g.Seats[0], g.Seats[1]
		pickBoard(g, other, 3)
		queueTheirPick(t, g, seat.ID, other.ID, 1, 1)
		moves := legal.EnumerateFor(g, seat.ID)
		assertOnlyChoiceMoves(t, g, seat.ID, moves)
		// The label says whose board is on offer — a run of these is
		// one prompt per player.
		if !anyLabelContains(moves, other.Name) {
			t.Errorf("no answer names the seat whose board is on offer: %v", labels(moves))
		}
	})
	t.Run("reveal_pick", func(t *testing.T) {
		g := newTable(t)
		seat, other := g.Seats[0], g.Seats[1]
		var revealed []uuid.UUID
		g.WithWriteLock(func() {
			for i := 0; i < 3 && i < len(other.Hand.Cards); i++ {
				revealed = append(revealed, other.Hand.Cards[i].InstanceID)
			}
			g.RevealForEffect(game.RevealSpec{Player: other.ID, Reason: "test", Cards: revealed})
			if _, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
				Chooser:  seat.ID,
				Owner:    other.ID,
				Question: "test — choose one of those cards",
				Cards:    revealed,
				Min:      1,
				Max:      1,
			}, func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil }); err != nil {
				t.Fatalf("RevealPickThenForEffect: %v", err)
			}
		})
		moves := legal.EnumerateFor(g, seat.ID)
		assertOnlyChoiceMoves(t, g, seat.ID, moves)
	})
}

func assertOnlyChoiceMoves(t *testing.T, g *game.Game, seat uuid.UUID, moves []legal.Move) {
	t.Helper()
	if len(moves) == 0 {
		t.Fatal("the seat owing the prompt was offered nothing at all — the #499 / #618 wedge")
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the seat owing the prompt was offered %q as well", m.Label)
		}
	}
	// Every answer the enumerator offers has to be one the engine
	// accepts — this package's one promise (#544).
	dispatchAll(t, g, seat, moves)
}

// choiceCardIDs is the single card each answer names, in offer order —
// the move list read as "which permanents reached the cap".
func choiceCardIDs(t *testing.T, moves []legal.Move) []string {
	t.Helper()
	var out []string
	for _, m := range moves {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal %q: %v", m.Label, err)
		}
		if len(p.CardIDs) == 1 {
			out = append(out, p.CardIDs[0])
		}
	}
	return out
}

func anyLabelContains(moves []legal.Move, want string) bool {
	for _, m := range moves {
		if strings.Contains(m.Label, want) {
			return true
		}
	}
	return false
}

// TestOwnPermanentsPickSpendsTheCheapestFuelFirst — own_permanents
// ranks the seat's OWN board by what it would miss least, which is
// Options.OrderCostFuel's question. With a board bigger than the
// expansion cap, the permanents that reach the move list are the ones
// the seat values LEAST.
func TestOwnPermanentsPickSpendsTheCheapestFuelFirst(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[0]
	all, biggest := pickBoard(g, seat, 20)
	queueOwnPick(t, g, seat.ID, 1, 1)
	keep := powerScorer(g)

	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{
		OrderCostFuel: func(c legal.TargetCandidate) float64 { return keep(c) },
	})
	offered := choiceCardIDs(t, moves)
	if len(offered) == 0 {
		t.Fatalf("nothing was offered: %v", labels(moves))
	}
	if len(offered) >= len(all) {
		t.Fatalf("the cap is not measuring anything (%d offers over %d permanents)", len(offered), len(all))
	}
	if contains(offered, biggest) {
		t.Errorf("the seat's best permanent reached the move list: the hook was read the wrong way round")
	}
	if offered[0] != all[0].String() {
		t.Errorf("the first offered sacrifice is %s, want the 1/1 %s", offered[0], all[0])
	}
	dispatchAll(t, g, seat.ID, moves)
}

// TestTheirPermanentsPickKeepsTheBiggestThreatInsideTheCap — the other
// direction. Choosing among somebody ELSE's board is
// Options.OrderTargets' question, so the part of it that matters is
// what survives the cap.
func TestTheirPermanentsPickKeepsTheBiggestThreatInsideTheCap(t *testing.T) {
	g := newTable(t)
	seat, other := g.Seats[0], g.Seats[1]
	all, biggest := pickBoard(g, other, 20)
	queueTheirPick(t, g, seat.ID, other.ID, 1, 1)

	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{
		OrderTargets: powerScorer(g),
	})
	offered := choiceCardIDs(t, moves)
	if len(offered) == 0 {
		t.Fatalf("nothing was offered: %v", labels(moves))
	}
	if len(offered) >= len(all) {
		t.Fatalf("the cap is not measuring anything (%d offers over %d permanents)", len(offered), len(all))
	}
	if !contains(offered, biggest) {
		t.Errorf("the biggest permanent is not among the %d offers — the cap dropped the one that matters", len(offered))
	}
	if offered[0] != biggest.String() {
		t.Errorf("the first offer is %s, want the 20/20 %s", offered[0], biggest)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// TestResolutionPickOrderingChangesOrderNotLegality — the hooks may not
// add or remove an answer, only decide which reach the cap. With a cap
// big enough for the whole board, hooked and unhooked enumerations
// offer the same set.
func TestResolutionPickOrderingChangesOrderNotLegality(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[0]
	pickBoard(g, seat, 6)
	queueOwnPick(t, g, seat.ID, 1, 1)

	plain := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{MaxExpansionPerSource: 50})
	hooked := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{
		MaxExpansionPerSource: 50,
		OrderCostFuel:         powerScorer(g),
	})
	a, b := choiceCardIDs(t, plain), choiceCardIDs(t, hooked)
	if len(a) != len(b) {
		t.Fatalf("the hook changed WHICH answers are legal: %d vs %d", len(a), len(b))
	}
	seen := map[string]bool{}
	for _, id := range a {
		seen[id] = true
	}
	for _, id := range b {
		if !seen[id] {
			t.Errorf("the hooked enumeration offered %s, which the plain one did not", id)
		}
	}
}
