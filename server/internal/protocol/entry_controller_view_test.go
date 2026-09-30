package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_controller_view_test.go — ADR 0102's wire: the choice gets a
// narrated line (owner decision 7), and the prompt projects its seats
// and its purpose.

// TestLogNarratesTheEntryControllerChoice: "Alice chose Bob to control
// Captive Audience." The chosen seat rides TargetSeat, like every
// other player reference in the log, and never `target`.
func TestLogNarratesTheEntryControllerChoice(t *testing.T) {
	g := buildActiveGame(t)
	chooser, chosen := g.Seats[0], g.Seats[1]
	card := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: card, Name: "Captive Audience", TypeLine: "Enchantment",
			Owner: chooser.ID, Controller: chosen.ID,
			KnownBy: map[uuid.UUID]bool{chooser.ID: true, chosen.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventEntryControllerChosen, Actor: chooser.ID, CardID: card,
			Target: chosen.ID, Label: "P2",
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogChooseController)
	if want := "P1 chose P2 to control Captive Audience"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.TargetSeat == nil || *entry.TargetSeat != 1 {
		t.Errorf("target_seat: got %v, want seat 1", entry.TargetSeat)
	}
	if entry.Target != "" {
		t.Errorf("target: got %q, want empty — the chosen controller is a seat", entry.Target)
	}
	assertNoUUID(t, entry.Text)
}

// TestEntryControllerPromptProjectsSeatsAndPurpose: the prompt's
// options go out as pick_options with `player` set, and its
// control_purpose beside them.
func TestEntryControllerPromptProjectsSeatsAndPurpose(t *testing.T) {
	g := buildActiveGame(t)
	// buildActiveGame seats two players; the second option is a seat
	// id the projection carries through without looking it up.
	chooser, a := g.Seats[0], g.Seats[1]
	b := struct{ ID uuid.UUID }{uuid.New()}
	g.WithWriteLock(func() {
		g.PendingChoices = append(g.PendingChoices, &game.PendingChoice{
			ID:      uuid.New(),
			Kind:    game.PendingChoiceEntryController,
			Chooser: chooser.ID,
			Reason:  "Captive Audience — choose an opponent to control it",
			PickOptions: []game.ChoiceOption{
				{Label: "P2", Player: a.ID},
				{Label: "P3", Player: b.ID},
			},
			ControlPurpose: game.ControlForHarm,
		})
	})
	v := FilterViewFor(ViewOfGame(g), chooser.ID.String())
	var got *PendingChoiceView
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceEntryController) {
			got = &v.PendingChoices[i]
		}
	}
	if got == nil {
		t.Fatal("the entry_controller prompt did not reach its chooser")
	}
	if got.ControlPurpose != "harm" {
		t.Errorf("control_purpose = %q, want harm", got.ControlPurpose)
	}
	if len(got.PickOptions) != 2 || got.PickOptions[0].Player != a.ID.String() || got.PickOptions[1].Player != b.ID.String() {
		t.Errorf("pick_options = %+v, want the two seats in order", got.PickOptions)
	}
}
