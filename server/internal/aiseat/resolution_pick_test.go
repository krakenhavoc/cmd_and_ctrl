package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// resolution_pick_test.go — the bot half of #1214, in the shape of
// choose_cards_validate_test.go.
//
// Three new PendingChoiceKinds are three new ways to build #544's
// ingredient list: a seat owing a blocking prompt is offered that
// prompt's answers AND NOTHING ELSE, so a kind `legal` cannot enumerate
// is a bot asleep on a table nobody else may move. The enumeration test
// in internal/legal catches a kind with no CASE; only a real run
// catches a case whose answers the engine refuses.
//
// So each kind is queued on a real table and the real heuristic plays
// every seat until the table is two turns past the prompt. A refusal of
// an answer to the prompt under test is an immediate failure: the
// enumerator promised it.
//
// Deliberately NOT behind requireGameTests, for the same reason the
// validated choose_cards run is not: it is a bounded loop, and the gate
// is how #544 shipped.

func TestHeuristicAnswersEveryResolutionPickAndTheTableAdvances(t *testing.T) {
	for _, tc := range []struct {
		name string
		// queue puts the prompt up and returns how many times its
		// continuation has run, by closing over a counter.
		queue func(t *testing.T, g *game.Game, seat *game.Player, other *game.Player, answers *int)
	}{
		{
			// "Target opponent chooses two of those cards" — the
			// asked seat is NOT the owner of the pool.
			name: "reveal_pick",
			queue: func(t *testing.T, g *game.Game, seat, other *game.Player, answers *int) {
				t.Helper()
				g.WithWriteLock(func() {
					if err := g.DrawNForEffect(seat.ID, 6); err != nil {
						t.Fatalf("draw: %v", err)
					}
					var revealed []uuid.UUID
					for _, c := range seat.Hand.Cards {
						revealed = append(revealed, c.InstanceID)
					}
					g.RevealForEffect(game.RevealSpec{
						Player: seat.ID,
						Reason: "test — reveal the hand",
						Cards:  revealed,
					})
					if _, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
						Chooser:  other.ID,
						Owner:    seat.ID,
						Question: "test — choose two of those cards",
						Cards:    revealed,
						Min:      2,
						Max:      2,
					}, func(*game.Game, []uuid.UUID, []uuid.UUID) error {
						*answers++
						return nil
					}); err != nil {
						t.Fatalf("RevealPickThenForEffect: %v", err)
					}
				})
			},
		},
		{
			// "You choose from among the permanents that player
			// controls" — one prompt per seat, answered by one seat.
			name: "their_permanents",
			queue: func(t *testing.T, g *game.Game, seat, other *game.Player, answers *int) {
				t.Helper()
				botPickTestBoard(g, other.ID, 3)
				g.WithWriteLock(func() {
					if _, err := g.PermanentsPickedThenForEffect(game.PermanentPickPrompt{
						Chooser:    seat.ID,
						Question:   "test — choose a permanent that player controls",
						Of:         []uuid.UUID{other.ID},
						Candidates: botPickCandidates(1, 1),
					}, func(*game.Game, game.PromptedPicks) error {
						*answers++
						return nil
					}); err != nil {
						t.Fatalf("PermanentsPickedThenForEffect: %v", err)
					}
				})
			},
		},
		{
			// "Sacrifice any number of lands" — the untargeted
			// self-choice, floor zero.
			name: "own_permanents",
			queue: func(t *testing.T, g *game.Game, seat, _ *game.Player, answers *int) {
				t.Helper()
				botPickTestBoard(g, seat.ID, 3)
				g.WithWriteLock(func() {
					if _, err := g.PermanentsPickedThenForEffect(game.PermanentPickPrompt{
						Chooser:    seat.ID,
						Question:   "test — sacrifice any number of permanents",
						Of:         []uuid.UUID{seat.ID},
						Candidates: botPickCandidates(0, 0),
					}, func(*game.Game, game.PromptedPicks) error {
						*answers++
						return nil
					}); err != nil {
						t.Fatalf("PermanentsPickedThenForEffect: %v", err)
					}
				})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolutionPickBotRun(t, tc.queue)
		})
	}
}

// botPickTestBoard puts n plain creatures on the battlefield under a
// seat's control, so a permanent pick has something to offer.
func botPickTestBoard(g *game.Game, controller uuid.UUID, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			g.Battlefield.PushTop(game.Card{
				InstanceID: uuid.New(),
				Name:       "Test Bear",
				TypeLine:   "Creature — Bear",
				Power:      2,
				Toughness:  2,
				Owner:      controller,
				Controller: controller,
			})
		}
	})
}

// botPickCandidates is every permanent a seat controls, with fixed
// bounds — the candidate function a card supplies.
func botPickCandidates(lo, hi int) func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
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

// resolutionPickBotRun queues one prompt and plays the whole table with
// the real heuristic until it is two turns past it.
func resolutionPickBotRun(t *testing.T, queue func(*testing.T, *game.Game, *game.Player, *game.Player, *int)) {
	g := newSettledTable(t, 22)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	answers := 0
	queue(t, g, seat, other, &answers)
	if len(g.PendingChoices) == 0 {
		t.Fatal("setup: no prompt was queued")
	}
	promptIDs := map[uuid.UUID]bool{}
	for _, c := range g.PendingChoices {
		if c != nil {
			promptIDs[c.ID] = true
		}
	}

	p := heuristic.New()
	rejects := map[uuid.UUID]int{}
	startTurn := g.Turn.Number
	for step := 1; step <= 600; step++ {
		if g.Turn.Number >= startTurn+2 {
			if answers != 1 {
				t.Fatalf("the table advanced but the prompt's continuation ran %d times", answers)
			}
			t.Logf("step %d: the prompt was answered and the table advanced two turns", step)
			return
		}
		var actor uuid.UUID
		for _, s := range g.Seats {
			if len(legal.EnumerateFor(g, s.ID)) > 0 {
				actor = s.ID
				break
			}
		}
		if actor == uuid.Nil {
			t.Fatalf("WEDGE at step %d: NO seat has a legal move. turn=%+v choices=%s",
				step, g.Turn, describeChoices(g))
		}
		moves := legal.EnumerateFor(g, actor)
		in := aiseat.Input{View: protocol.ViewOfGameFor(g, actor.String()), Seat: actor, Moves: moves}
		d, err := p.Decide(context.Background(), in)
		if err != nil {
			t.Fatalf("Decide: %v", err)
		}
		if d.Index == aiseat.Decline {
			pi := aiseat.PassIndex(moves)
			if pi < 0 {
				t.Fatalf("WEDGE at step %d: the policy declines and there is no pass. choices=%s", step, describeChoices(g))
			}
			d.Index = pi
		}
		chosen := moves[d.Index]
		derr := actions.Dispatch(g, actions.Action{
			Type: actions.Type(chosen.Type), Player: actor, Caller: actor, Params: chosen.Params,
		})
		if derr == nil {
			rejects[actor] = 0
			continue
		}
		// A refusal while a prompt under test is still open is the
		// bug this run exists to catch: the enumerator promised the
		// answer it just offered.
		for _, c := range g.PendingChoices {
			if c != nil && promptIDs[c.ID] {
				t.Fatalf("step %d: the engine refused the enumerated answer %q: %v", step, chosen.Label, derr)
			}
		}
		rejects[actor]++
		if aiseat.PassIndex(moves) < 0 && rejects[actor] >= 3 {
			t.Fatalf("WEDGE at step %d: %q is rejected every time and no pass is on offer. choices=%s",
				step, chosen.Label, describeChoices(g))
		}
	}
	t.Fatalf("WEDGE: 600 bot steps without the table advancing two turns. answers=%d turn=%+v choices=%s",
		answers, g.Turn, describeChoices(g))
}
