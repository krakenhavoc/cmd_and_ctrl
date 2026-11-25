package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// color_choice_multi_test.go pins #6146: "becomes the color or colors of
// your choice" is a choose_color prompt whose answer is a non-empty SET of
// colours (CR 105.3: the new colours replace the old ones; an effect that
// gives an object colours never leaves it with none).

func TestColorChoiceMultiAcceptsASetAndRefusesTheEmptyOne(t *testing.T) {
	g := newActiveGame(t)
	seat, other := g.Seats[0].ID, g.Seats[1].ID
	resetProbes()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueColorChoiceThenForEffect(ColorPrompt{
			Chooser: seat, Question: "Quickchange — choose colors", Purpose: ColorForBenefit,
			Multi: true, ThenTail: probeTail(),
		})
	})
	c := choiceByKind(g, PendingChoiceColor)
	if c == nil || !c.ColorMulti || len(c.ColorOptions) != 5 {
		t.Fatalf("queued choice = %+v, want a multi pick of the five colours", c)
	}
	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Error("an open multi colour pick is not a restore point")
	}
	for _, bad := range []string{"", " , ", "WC", "X", "WUBRGC"} {
		if err := g.ResolveColorChoice(id, seat, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("answer %q: err = %v, want ErrInvalidParam (the prompt stays open)", bad, err)
		}
	}
	if err := g.ResolveColorChoice(id, other, "W"); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat answered: err = %v, want ErrNotTheChooser", err)
	}
	// Any order, any case, repeats and separators: the set is WUBRG-ordered.
	if err := g.ResolveColorChoice(id, seat, "u, w,U"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	if len(testProbeHeard) != 1 || testProbeHeard[0].Word != "WU" {
		t.Errorf("continuation was told %+v, want WU once", testProbeHeard)
	}
	var label string
	for _, ev := range g.Events {
		if ev.Kind == EventColorChosen {
			label = ev.Label
		}
	}
	if label != "WU" {
		t.Errorf("color_chosen label = %q, want WU", label)
	}
}

// One colour is a legal set, so a client that renders the prompt as an
// ordinary colour picker still answers it.
func TestColorChoiceMultiAcceptsOneColour(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	resetProbes()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueColorChoiceThenForEffect(ColorPrompt{Chooser: seat, Multi: true, Purpose: ColorForBenefit, ThenTail: probeTail()})
	})
	if err := g.ResolveColorChoice(id, seat, "g"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	if len(testProbeHeard) != 1 || testProbeHeard[0].Word != "G" {
		t.Errorf("continuation was told %+v, want G", testProbeHeard)
	}
}

func TestColorChoiceMultiSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.QueueColorChoiceThenForEffect(ColorPrompt{Chooser: seat, Multi: true, Purpose: ColorForBenefit, ThenTail: probeTail()})
	})
	restored := restoreStrictForTest(t, g)
	c := choiceByKind(restored, PendingChoiceColor)
	if c == nil || !c.ColorMulti {
		t.Fatalf("restored choice = %+v, want the multi flag kept", c)
	}
}

func TestColorSetHelpers(t *testing.T) {
	if got := ColorSetLetters("gw"); len(got) != 2 || got[0] != "W" || got[1] != "G" {
		t.Errorf("ColorSetLetters(gw) = %v, want [W G]", got)
	}
	for word, want := range map[string]string{"U": "blue", "WU": "white and blue", "WUB": "white, blue and black", "": ""} {
		if got := ColorSetName(word); got != want {
			t.Errorf("ColorSetName(%q) = %q, want %q", word, got, want)
		}
	}
}
