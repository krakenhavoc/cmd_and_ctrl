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

// choose_cards_validate_test.go — the bot half of #624, in the shape of
// myriad_wedge_test.go.
//
// A ChooseCardsPrompt may now carry a set-level Validate hook, and that
// is precisely the ingredient list of #544: a rule on an unexported
// frame, a resolver that rejects without dequeuing so a human can try
// again, and a seat owing a choice that is offered nothing but that
// choice's answers. A deterministic bot handed a set the rule refuses
// retries it forever; a bot handed an empty list sleeps. Either way the
// table stops.
//
// So this runs the real heuristic through the real engine, every seat,
// until the table is two turns past the prompt, and fails on either
// shape. The prompt is a test-only "discard two cards unless you
// discard a creature card" (the Invasion of New Phyrexia +1 this hook
// exists for is not registered here; that is #626). The hand is drawn
// deep enough that the candidate count clears MaxExpansionPerSource,
// which is where a smallest-first enumerator runs out of budget on
// singles the rule rejects and never reaches a pair.
//
// Deliberately NOT behind requireGameTests, for chained_choice_test.go's
// reason: it is a bounded loop, and the gate is how #544 shipped.

func TestHeuristicAnswersAValidatedChooseCardsPromptAndTheTableAdvances(t *testing.T) {
	for _, tc := range []struct {
		name string
		// candidates picks the prompt's card list out of the hand.
		candidates func(hand []game.Card) []uuid.UUID
	}{
		{
			// Every single is invalid: only pairs can answer.
			name: "no_creature_among_candidates",
			candidates: func(hand []game.Card) []uuid.UUID {
				var out []uuid.UUID
				for _, c := range hand {
					if !c.IsCreature() {
						out = append(out, c.InstanceID)
					}
				}
				return out
			},
		},
		{
			name: "whole_hand",
			candidates: func(hand []game.Card) []uuid.UUID {
				out := make([]uuid.UUID, 0, len(hand))
				for _, c := range hand {
					out = append(out, c.InstanceID)
				}
				return out
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			validatedChooseCardsRun(t, tc.candidates)
		})
	}
}

func validatedChooseCardsRun(t *testing.T, pickCandidates func([]game.Card) []uuid.UUID) {
	g := newSettledTable(t, 21)
	seat := g.Seats[g.Turn.ActiveSeat]

	var (
		cands   []uuid.UUID
		answers int
		taken   []uuid.UUID
	)
	g.WithWriteLock(func() {
		// Deep enough that the non-creature candidates alone clear the
		// expansion budget of 12.
		if err := g.DrawNForEffect(seat.ID, 20); err != nil {
			t.Fatalf("draw: %v", err)
		}
		cands = pickCandidates(seat.Hand.Cards)
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  seat.ID,
			Question: "Discard two cards unless you discard a creature card",
			Cards:    cands,
			Min:      1,
			Max:      2,
			Zone:     game.ZoneHand,
			Validate: func(picked []game.Card) bool {
				return len(picked) == 2 || (len(picked) == 1 && picked[0].IsCreature())
			},
			Then: func(g *game.Game, picked []uuid.UUID) error {
				answers++
				taken = picked
				p := g.PlayerByIDForEffect(seat.ID)
				for _, id := range picked {
					if _, err := game.MoveCard(p.Hand, p.Graveyard, id); err != nil {
						return err
					}
				}
				return nil
			},
		})
	})
	// 12 is legal's default MaxExpansionPerSource.
	if len(cands) < 12 {
		t.Fatalf("setup: %d candidates does not clear the expansion budget", len(cands))
	}
	promptID := g.PendingChoices[0].ID

	p := heuristic.New()
	rejects := map[uuid.UUID]int{}
	startTurn := g.Turn.Number
	for step := 1; step <= 600; step++ {
		if g.Turn.Number >= startTurn+2 {
			if answers != 1 {
				t.Fatalf("the table advanced but the prompt's continuation ran %d times", answers)
			}
			t.Logf("step %d: prompt answered with %d card(s) over %d candidates; table advanced two turns",
				step, len(taken), len(cands))
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
		// Any refusal of an answer to THIS prompt is the bug: the
		// enumerator promised it.
		for _, c := range g.PendingChoices {
			if c != nil && c.ID == promptID {
				t.Fatalf("step %d: the engine refused the enumerated answer %q: %v — the enumerator offered a set the prompt's Validate rejects",
					step, chosen.Label, derr)
			}
		}
		rejects[actor]++
		t.Logf("step %d: engine rejected %q for %s: %v", step, chosen.Label, actor, derr)
		if aiseat.PassIndex(moves) < 0 && rejects[actor] >= 3 {
			t.Fatalf("WEDGE at step %d: %q is rejected every time and no pass is on offer. choices=%s",
				step, chosen.Label, describeChoices(g))
		}
	}
	t.Fatalf("WEDGE: 600 bot steps without the table advancing two turns. answers=%d turn=%+v choices=%s",
		answers, g.Turn, describeChoices(g))
}
