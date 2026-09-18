package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// leaving_the_game_test.go — #902, the bot-arena half of CR 800.4g.
//
// The S36 symptom this issue exists for is a table that wedges: a seat
// concedes while a prompt addressed to it is open, and the other three
// seats sit behind a question nobody can answer. #868 stopped that by
// DROPPING the prompt; CR 800.4g says a prompt about somebody else's
// object is reassigned instead. Either way the table has to keep
// playing, and this is the test that says so from the runner's side —
// bots only, no human to unstick anything.
//
// Four seats, because a concede in a two-player game ends the game
// before anything can inherit anything (ADR 0060 Decision 5).

func departureBotDeck(seat int) []game.Card {
	deck := make([]game.Card, 0, 61)
	cmdr := game.NewCommander(fmt.Sprintf("Commander %d", seat), uuid.Nil)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{G}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	for i := 0; i < 40; i++ {
		deck = append(deck, game.NewCard("Forest", uuid.Nil))
	}
	for i := 0; i < 20; i++ {
		c := game.NewCard(fmt.Sprintf("Bear %d-%d", seat, i), uuid.Nil)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{G}"
		c.Power, c.Toughness = 2, 2
		deck = append(deck, c)
	}
	for i := range deck {
		if deck[i].TypeLine == "" {
			deck[i].TypeLine = "Basic Land — Forest"
		}
	}
	return deck
}

// TestConcedeMidPromptLeavesTheBotTablePlaying opens Fact or Fiction's
// pile-split prompt on seat 1, concedes seat 1 while it is open, and
// then drives every seat with the heuristic policy the arena uses. The
// prompt must reach a seat that can answer it (CR 800.4g) and the table
// must clear it and move on.
func TestConcedeMidPromptLeavesTheBotTablePlaying(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), departureBotDeck(i)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 13))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	caster, leaver := g.Seats[0], g.Seats[1]

	source := game.NewCard("Fact or Fiction", caster.ID)
	source.TypeLine = "Instant"
	source.InstanceID = uuid.New()
	source.Controller = caster.ID
	g.Battlefield.PushTop(source)

	if caster.Hand == nil || len(caster.Hand.Cards) < 3 {
		t.Fatalf("setup: the caster holds %d cards, need 3", len(caster.Hand.Cards))
	}
	cards := []uuid.UUID{
		caster.Hand.Cards[0].InstanceID,
		caster.Hand.Cards[1].InstanceID,
		caster.Hand.Cards[2].InstanceID,
	}
	split := 0
	g.WithWriteLock(func() {
		g.QueuePileSplitForEffect(game.PileSplitPrompt{
			Splitter:      leaver.ID,
			Chooser:       caster.ID,
			Owner:         caster.ID,
			Source:        source.InstanceID,
			SplitQuestion: "Separate those cards into two piles",
			PickQuestion:  "Take a pile",
			Cards:         cards,
			Then: func(*game.Game, []uuid.UUID, []uuid.UUID) error {
				split++
				return nil
			},
		})
	})
	if len(legal.EnumerateFor(g, leaver.ID)) == 0 {
		t.Fatal("setup: the splitter was offered no answers before conceding")
	}

	// The concede lands with the prompt open — the exact window the
	// S31 arena kept wedging in.
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	// The prompt must have MOVED, not vanished — otherwise the loop
	// below would pass for the wrong reason (#868's drop also leaves a
	// playable table).
	moved := uuid.Nil
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				moved = c.Chooser
			}
		}
	})
	if moved == uuid.Nil {
		t.Fatalf("the split prompt was dropped rather than reassigned (CR 800.4g)")
	}
	if moved == leaver.ID {
		t.Fatalf("the split prompt is still addressed to the seat that left")
	}
	t.Logf("the split prompt moved from %s to %s", leaver.ID, moved)

	p := heuristic.New()
	for step := 1; step <= 200; step++ {
		if split > 0 {
			return
		}
		var actor uuid.UUID
		for _, s := range g.Seats {
			if s.Eliminated {
				continue
			}
			if len(legal.EnumerateFor(g, s.ID)) > 0 {
				actor = s.ID
				break
			}
		}
		if actor == uuid.Nil {
			t.Fatalf("WEDGE at step %d: no seat still in the game has a legal move. turn=%+v choices=%s",
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
				t.Fatalf("WEDGE at step %d: the policy declines and there is no pass. choices=%s",
					step, describeChoices(g))
			}
			d.Index = pi
		}
		chosen := moves[d.Index]
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.Type(chosen.Type), Player: actor, Caller: actor, Params: chosen.Params,
		}); err != nil {
			t.Logf("step %d: engine rejected %q for %s: %v", step, chosen.Label, actor, err)
		}
	}
	t.Fatalf("WEDGE: 200 bot steps and the reassigned prompt was never resolved. choices=%s", describeChoices(g))
}
