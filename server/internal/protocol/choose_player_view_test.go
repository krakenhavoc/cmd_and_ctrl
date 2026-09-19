package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_player_view_test.go — #929 on the wire.
//
// A choose-a-player prompt is the one Options-bearing prompt with
// NOTHING to hide: its branches are seats, and who is seated is public
// (CR 400.2). So the assertion here is the mirror image of
// option_pick_view_test.go's — every viewer sees the same seat names,
// the redaction pass drops nothing, and no card IDs ride along that a
// viewer could correlate later.

func TestChoosePlayerPromptShowsEverySeatToEveryViewer(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
			Chooser:  chooser.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID},
			Question: "Slithermuse — choose an opponent",
			Item:     &game.StackItem{ID: uuid.New(), Controller: chooser.ID},
			Then:     func(*game.Game, uuid.UUID) error { return nil },
		})
	})

	for _, viewer := range g.Seats {
		got := FilterViewFor(ViewOfGame(g), viewer.ID.String())
		c := choiceFor(t, got, string(game.PendingChoiceOptionPick))
		if len(c.PickOptions) != 2 {
			t.Fatalf("viewer %s sees %d options, want both seats", viewer.Name, len(c.PickOptions))
		}
		for i, want := range []string{g.Seats[1].Name, g.Seats[2].Name} {
			if c.PickOptions[i].Label != want {
				t.Errorf("viewer %s option %d = %q, want %q",
					viewer.Name, i, c.PickOptions[i].Label, want)
			}
			if len(c.PickOptions[i].Cards) != 0 {
				t.Errorf("a seat option carries no cards: %+v", c.PickOptions[i])
			}
		}
		if c.Chooser != chooser.ID.String() {
			t.Errorf("the prompt names its chooser: %q", c.Chooser)
		}
	}
}
