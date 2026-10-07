package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0082's 2026-10-07 amendment (#2570): manifest dread's choice is a
// "look at", so the pair is the looker's alone. The prompt carries the
// two cards for the chooser and nothing at all for the other seat, and
// the finished action reads on the public log without naming a card.
func TestManifestDreadPromptIsPrivateToItsController(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.ManifestDreadThenForEffect(me.ID, uuid.Nil, nil); err != nil {
			t.Fatalf("ManifestDreadThenForEffect: %v", err)
		}
	})

	own := FilterViewFor(ViewOfGame(g), me.ID.String()).PendingChoices
	if len(own) != 1 || len(own[0].Options) != 2 {
		t.Fatalf("controller sees %d prompts, want one over 2 candidates", len(own))
	}
	if own[0].ChooseMin != 1 || own[0].ChooseMax != 1 {
		t.Errorf("prompt bounds %d..%d, want exactly one", own[0].ChooseMin, own[0].ChooseMax)
	}
	other := FilterViewFor(ViewOfGame(g), them.ID.String()).PendingChoices
	if len(other) != 1 || len(other[0].Options) != 0 {
		t.Errorf("the other seat sees %d prompts with %d candidates, want the prompt without its cards", len(other), len(other[0].Options))
	}
}

func TestManifestDreadLogLineNamesNoCard(t *testing.T) {
	turn, step := 1, ""
	var sacrificed uuid.UUID
	e, ok := projectEvent(game.Event{Kind: game.EventManifestDread, Actor: uuid.New(), CardID: uuid.New(), Target: uuid.New()},
		func(uuid.UUID) int { return 0 }, &turn, &step, &sacrificed)
	if !ok || e.Kind != LogManifestDread {
		t.Fatalf("projected %+v, %v", e, ok)
	}
	if e.CardID != "" {
		t.Errorf("the line carries card %q; the manifested card is hidden", e.CardID)
	}
	e.actorName = "Alice"
	if got := renderLogText(e, "", ""); got != "Alice manifested dread" {
		t.Errorf("line = %q", got)
	}
}
