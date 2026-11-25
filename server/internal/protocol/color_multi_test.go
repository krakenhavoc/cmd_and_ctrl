package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// color_multi_test.go — #6146: "the color or colors of your choice" rides
// the wire as color_multi on its choose_color prompt, and its answer logs
// as a set.

func TestColorMultiPromptIsProjected(t *testing.T) {
	g := buildActiveGame(t)
	chooser := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueColorChoiceThenForEffect(game.ColorPrompt{
			Chooser: chooser.ID, Question: "Scuttlemutt — choose one or more colors",
			Purpose: game.ColorForBenefit, Multi: true,
		})
	})
	var found bool
	for _, c := range ViewOfGame(g).PendingChoices {
		if c.Kind == string(game.PendingChoiceColor) {
			found = true
			if !c.ColorMulti || len(c.ColorOptions) != 5 {
				t.Errorf("projected prompt = %+v, want color_multi over five colours", c)
			}
		}
	}
	if !found {
		t.Fatal("no choose_color prompt on the wire")
	}
}

func TestLogNarratesAChosenColorSet(t *testing.T) {
	g := buildActiveGame(t)
	chooser, other := g.Seats[0], g.Seats[1]
	mutt := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: mutt, Name: "Scuttlemutt", TypeLine: "Artifact Creature — Scarecrow",
			Owner: chooser.ID, Controller: chooser.ID,
			KnownBy: map[uuid.UUID]bool{chooser.ID: true, other.ID: true},
		})
		g.EmitEvent(game.Event{Kind: game.EventColorChosen, Actor: chooser.ID, CardID: mutt, Label: "WU"})
	})
	entry := findLog(t, ViewOfGame(g).Log, LogChooseColor)
	if want := "P1 chose white and blue for Scuttlemutt"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
}
