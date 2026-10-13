package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// villainous_choice_test.go — the enumerator half of #3140.
//
// A villainous choice (CR 701.55) is an option pick addressed to the
// player who faces it, with both options bad. CR 701.55b lets a player
// choose an option that is illegal or impossible and perform as much of
// it as they can, so the enumerator must offer BOTH, whatever the
// chooser's board is. A seat that was shown only the options it could
// fully carry out would be shown one, or none, and a prompt with no
// answers is a wedged table (#544).

// queueVillainous parks the prompt the way effects.FaceVillainousChoice
// lays it out: the facing player is the chooser, not the controller.
func queueVillainous(g *game.Game, facing uuid.UUID, a, b game.ChoiceOption) {
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  facing,
			Question: "Damocles Base — sacrifice a nontoken creature, or lose 2 life?",
			Options:  []game.ChoiceOption{a, b},
			Then:     func(*game.Game, int) error { return nil },
		})
	})
}

func TestAVillainousChoiceOffersBothOptionsToTheFacingPlayerOnly(t *testing.T) {
	g := newTable(t)
	me, facing := g.Seats[0], g.Seats[1]
	queueVillainous(g, facing.ID,
		game.ChoiceOption{Label: "Sacrifice a nontoken creature", Harm: 2},
		game.ChoiceOption{Label: "Lose 2 life", LifeCost: 2, Harm: 1})

	moves := legal.EnumerateFor(g, facing.ID)
	if len(moves) != 2 {
		t.Fatalf("enumerated %d answers, want both options: %v", len(moves), labels(moves))
	}
	for _, want := range []string{
		"Damocles Base — sacrifice a nontoken creature, or lose 2 life?: Sacrifice a nontoken creature",
		"Damocles Base — sacrifice a nontoken creature, or lose 2 life?: Lose 2 life",
	} {
		if !hasLabel(moves, want) {
			t.Errorf("missing %q: %v", want, labels(moves))
		}
	}
	for _, m := range moves {
		if m.Label == "Damocles Base — sacrifice a nontoken creature, or lose 2 life?: Lose 2 life" &&
			(m.Cost == nil || m.Cost.Life != 2) {
			t.Errorf("the life branch is priced for a bot: %+v", m.Cost)
		}
	}
	dispatchAll(t, g, facing.ID, moves)

	// Nothing is owed by the controller: the question was never theirs.
	g2 := newTable(t)
	queueVillainous(g2, g2.Seats[1].ID,
		game.ChoiceOption{Label: "A"}, game.ChoiceOption{Label: "B", Harm: 1})
	for _, m := range legal.EnumerateFor(g2, g2.Seats[0].ID) {
		if m.Kind == legal.KindChoice {
			t.Errorf("the controller (%s) was offered an answer to the opponent's question: %s", me.ID, m.Label)
		}
	}
}

// An option that would do nothing to this player (no creature to
// sacrifice) is still an answer the engine accepts.
func TestAVillainousChoiceStillOffersAnOptionThatWouldDoNothing(t *testing.T) {
	g := newTable(t)
	facing := g.Seats[1]
	queueVillainous(g, facing.ID,
		game.ChoiceOption{Label: "Sacrifice a nontoken creature"}, // Harm zeroed: nothing to sacrifice
		game.ChoiceOption{Label: "Lose 2 life", LifeCost: 2, Harm: 1})

	moves := legal.EnumerateFor(g, facing.ID)
	if len(moves) != 2 {
		t.Fatalf("enumerated %d answers, want 2: %v", len(moves), labels(moves))
	}
	always := 0
	for _, m := range moves {
		if m.AlwaysLegal {
			always++
		}
	}
	if always != 1 {
		t.Errorf("%d answers marked always-legal, want exactly the first", always)
	}
	dispatchAll(t, g, facing.ID, moves)
}
