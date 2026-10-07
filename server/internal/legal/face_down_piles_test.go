package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// face_down_piles_test.go — #2147. The bot's closed move list has to
// answer both legs of a face-down split: the separator's "which cards
// go face down" and the chooser's "which pile", where the chooser
// cannot see the face-down cards.

func TestFaceDownSplitIsEnumeratedAtBothSteps(t *testing.T) {
	g := newTable(t)
	caster, splitter := g.Seats[0], g.Seats[1]
	source := battlefieldCard(g, caster, game.Card{Name: "Sauron's Ransom", TypeLine: "Instant"})

	var looked []uuid.UUID
	g.WithWriteLock(func() {
		looked = g.LookAtTopOfPlayersLibraryForEffect(splitter.ID, caster.ID, 4)
		g.QueuePileSplitForEffect(game.PileSplitPrompt{
			Splitter:      splitter.ID,
			Chooser:       caster.ID,
			Owner:         caster.ID,
			Source:        source,
			SplitQuestion: "Choose the cards for the face-down pile",
			PickQuestion:  "Take a pile",
			Cards:         looked,
			FaceDown:      true,
			Then:          func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil },
		})
	})
	if len(looked) != 4 {
		t.Fatalf("setup: looked at %d cards", len(looked))
	}

	// Step one: the separator is offered answers, the caster nothing.
	split := legal.EnumerateFor(g, splitter.ID)
	if len(split) == 0 {
		t.Fatal("the separator was offered no answer to the split")
	}
	for _, m := range split {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the separator was offered %q as well as the split", m.Label)
		}
	}
	if n := len(legal.EnumerateFor(g, caster.ID)); n != 0 {
		t.Errorf("the caster was offered %d moves while the opponent separates", n)
	}
	dispatchAll(t, g, splitter.ID, split)

	// Answer it for real: any offered split will do.
	m := split[0]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: splitter.ID, Params: m.Params,
	}); err != nil {
		t.Fatalf("dispatch the split: %v", err)
	}

	// Step two: the caster picks a pile; every answer is accepted and
	// none of the labels names a card.
	pick := legal.EnumerateFor(g, caster.ID)
	if len(pick) != 2 {
		t.Fatalf("two piles to take, got %d: %v", len(pick), labels(pick))
	}
	dispatchAll(t, g, caster.ID, pick)
	var up, down bool
	for _, p := range pick {
		up = up || strings.Contains(p.Label, "face-up")
		down = down || strings.Contains(p.Label, "face-down")
		for _, id := range looked {
			c, _ := g.LookupCardForEffect(id)
			if c.Name != "" && strings.Contains(p.Label, c.Name) {
				t.Errorf("the pick's label %q names a card the chooser cannot see", p.Label)
			}
		}
	}
	if !up || !down {
		t.Errorf("the two answers are the face-up and the face-down pile: %v", labels(pick))
	}
	if n := len(legal.EnumerateFor(g, splitter.ID)); n != 0 {
		t.Errorf("the separator was offered %d moves while the caster picks", n)
	}
}
