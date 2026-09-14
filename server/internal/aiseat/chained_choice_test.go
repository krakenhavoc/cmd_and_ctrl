package aiseat_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// chained_choice_test.go — can a bot finish a chain?
//
// #544 is the reason this file exists. A seat that owes a choice is
// enumerated that choice's answers and NOTHING else, so a prompt the
// enumerator cannot answer is a sleeping bot holding a live table —
// which happened to a human opponent on turn 2. A chain multiplies the
// exposure: every link is another prompt that has to be answerable,
// and the last link is the one that has to actually END.
//
// So this walks the real policy through the real engine, one enumerated
// move at a time, and fails on the two shapes that wedge: an empty move
// list, and an answer the engine refuses.
//
// Deliberately NOT behind requireGameTests. It is not a whole game — it
// is a bounded loop over one card's prompts, in milliseconds — and the
// gate on the whole-game tests is precisely why #544 reached
// production.

// newSettledTable is newRoom past the mulligan window: every seat has
// kept, so the enumerator is out of its "keep or mulligan" world and
// back to offering real moves.
func newSettledTable(t *testing.T, seed uint64) *game.Game {
	t.Helper()
	g := newRoom(t, 2, seed).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// driveChoices runs the policy against the seat for as long as it owes
// a pending choice, dispatching each decision. Returns the labels it
// took. Fails on an empty move list (the #544 wedge) or a rejected
// dispatch (the #543 one).
func driveChoices(t *testing.T, g *game.Game, pol aiseat.Policy, seat uuid.UUID, maxSteps int) []string {
	t.Helper()
	var taken []string
	for step := 0; step < maxSteps; step++ {
		if !owesChoice(g, seat) {
			return taken
		}
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: the seat owes a choice and was offered NOTHING — "+
				"this is the #544 wedge: the runner sleeps and the table stops. "+
				"Open choices: %s", step, openChoices(g, seat))
		}
		in := aiseat.Input{
			View:  protocol.ViewOfGameFor(g, seat.String()),
			Seat:  seat,
			Moves: moves,
		}
		d, err := pol.Decide(context.Background(), in)
		if err != nil {
			t.Fatalf("step %d: policy failed: %v", step, err)
		}
		if d.Index == aiseat.Decline {
			t.Fatalf("step %d: the policy declined a window it owes an answer in. "+
				"A decline with no pass on offer returns the runner to sleep "+
				"(runner.go) — a chain link must be answerable, not skippable.", step)
		}
		if d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("step %d: policy returned out-of-range index %d of %d", step, d.Index, len(moves))
		}
		mv := moves[d.Index]
		if err := actions.Dispatch(g, actions.Action{
			Type:   actions.Type(mv.Type),
			Player: mv.Player,
			Caller: seat,
			Params: mv.Params,
		}); err != nil {
			t.Fatalf("step %d: the engine refused an ENUMERATED answer %q: %v — "+
				"an enumerator that offers what the resolver rejects is #544",
				step, mv.Label, err)
		}
		taken = append(taken, mv.Label)
	}
	t.Fatalf("the chain did not terminate in %d steps; took %v", maxSteps, taken)
	return nil
}

func owesChoice(g *game.Game, seat uuid.UUID) bool {
	owed := false
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == seat {
				owed = true
				return
			}
		}
	})
	return owed
}

func openChoices(g *game.Game, seat uuid.UUID) string {
	var b strings.Builder
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == seat {
				b.WriteString(string(c.Kind))
				b.WriteString("(" + c.Reason + ") ")
			}
		}
	})
	return b.String()
}

// TestHeuristicFinishesPondersShuffleChain — reorder, then the shuffle
// question, then done, with the draw landing either way.
func TestHeuristicFinishesPondersShuffleChain(t *testing.T) {
	g := newSettledTable(t, 4)
	seat := g.Seats[0]
	handBefore := seat.Hand.Size()

	g.WithWriteLock(func() {
		g.LookAtTopThenForEffect(seat.ID, uuid.Nil, 3, func(g *game.Game) error {
			g.QueueConfirmForEffect(game.ConfirmPrompt{
				Chooser:      seat.ID,
				Question:     "Ponder — shuffle your library?",
				AcceptLabel:  "Shuffle",
				DeclineLabel: "Keep that order",
				OnAccept: func(g *game.Game) error {
					if err := g.ShuffleLibraryForEffect(seat.ID); err != nil {
						return err
					}
					return g.DrawNForEffect(seat.ID, 1)
				},
				OnDecline: func(g *game.Game) error {
					return g.DrawNForEffect(seat.ID, 1)
				},
			})
			return nil
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 8)
	if len(took) != 2 {
		t.Errorf("took %d answers, want 2 (the reorder and the shuffle): %v", len(took), took)
	}
	if seat.Hand.Size() != handBefore+1 {
		t.Errorf("hand is %d, want %d — the draw at the end of the chain never ran",
			seat.Hand.Size(), handBefore+1)
	}
}

// TestHeuristicFinishesSylvanLibrarysThreeLinkChain is the whole card,
// bot-side: pick two cards, then answer a pay-or-put-back question per
// card, each queued by the answer before it.
func TestHeuristicFinishesSylvanLibrarysThreeLinkChain(t *testing.T) {
	g := newSettledTable(t, 9)
	seat := g.Seats[0]

	// Stand in for the trigger's resolution: draw two, then ask.
	var drawn []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(seat.ID, 2); err != nil {
			t.Fatalf("draw: %v", err)
		}
		drawn = g.CardsDrawnThisTurnFor(seat.ID)
	})
	if len(drawn) < 2 {
		t.Fatalf("seeded %d drawn cards, want at least 2", len(drawn))
	}
	handBefore := seat.Hand.Size()
	lifeBefore := seat.Life
	libBefore := seat.Library.Size()

	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  seat.ID,
			Question: "Sylvan Library — choose two cards drawn this turn",
			Cards:    drawn,
			Min:      2,
			Max:      2,
			Zone:     game.ZoneHand,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				return sylvanChainForTest(g, seat.ID, picked)
			},
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 12)
	// One pick plus one question per chosen card.
	if len(took) != 3 {
		t.Errorf("took %d answers, want 3: %v", len(took), took)
	}
	// Whatever the policy chose, the books balance: every card either
	// stayed in hand and cost 4 life, or went back to the library.
	paid := (lifeBefore - seat.Life) / 4
	putBack := seat.Library.Size() - libBefore
	if paid+putBack != 2 {
		t.Errorf("paid for %d and put back %d, want 2 cards accounted for", paid, putBack)
	}
	if seat.Hand.Size() != handBefore-putBack {
		t.Errorf("hand is %d, want %d", seat.Hand.Size(), handBefore-putBack)
	}
}

// TestHeuristicWillNotPayItsLastLifeToAChainedPrompt — the hole #547
// closed for activated abilities, reached through a prompt instead. A
// confirm declares what its accept branch charges, the enumerator
// stamps it on the move, and the policy prices it: a bot on 4 life
// answering "pay 4 life to keep this card" would be answering "lose the
// game", and it must take the other branch instead.
func TestHeuristicWillNotPayItsLastLifeToAChainedPrompt(t *testing.T) {
	g := newSettledTable(t, 11)
	seat := g.Seats[0]
	paid := false
	g.WithWriteLock(func() {
		seat.Life = 4
		g.QueueConfirmForEffect(game.ConfirmPrompt{
			Chooser:      seat.ID,
			Question:     "Sylvan Library — pay 4 life to keep it?",
			AcceptLabel:  "Pay 4 life",
			DeclineLabel: "Put it on top",
			LifeCost:     4,
			OnAccept: func(*game.Game) error {
				paid = true
				return nil
			},
		})
	})

	took := driveChoices(t, g, heuristic.New(), seat.ID, 4)
	if paid {
		t.Errorf("the bot paid its last 4 life; it took %v", took)
	}
	if len(took) != 1 || !strings.Contains(took[0], "Put it on top") {
		t.Errorf("took %v, want the decline branch", took)
	}
}

// sylvanChainForTest mirrors the card's per-card chain without
// importing the effects package (which would import legal's catalog
// hooks into this test's dependency graph for no benefit).
func sylvanChainForTest(g *game.Game, controller uuid.UUID, remaining []uuid.UUID) error {
	if len(remaining) == 0 {
		return nil
	}
	card, rest := remaining[0], remaining[1:]
	putBack := func(g *game.Game) error {
		if err := g.TuckToLibraryForEffect(card, false); err != nil {
			return err
		}
		return sylvanChainForTest(g, controller, rest)
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      controller,
		Question:     "Sylvan Library — pay 4 life to keep it?",
		AcceptLabel:  "Pay 4 life",
		DeclineLabel: "Put it on top",
		OnAccept: func(g *game.Game) error {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, controller, -4); err != nil {
				return err
			}
			return sylvanChainForTest(g, controller, rest)
		},
		OnDecline: putBack,
	})
	return nil
}
