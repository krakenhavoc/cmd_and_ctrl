package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_storm_test.go — #1238 / ADR 0086 Decision 5. The storm count is
// the only number on the card and the only thing that explains why N
// copies just appeared, and nothing else in the log says it: a copy
// of a spell is created rather than cast (CR 707.10) so it emits no
// event, and the trigger's own `resolve` entry carries no card.

// seedStormLog puts a storm spell on the stack and emits its count.
func seedStormLog(t *testing.T, g *game.Game, count int) uuid.UUID {
	t.Helper()
	caster := g.Seats[0]
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spell,
			Name:       "Grapeshot",
			TypeLine:   "Sorcery",
			Owner:      caster.ID,
			Controller: caster.ID,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventStorm, Actor: caster.ID, Source: spell, CardID: spell, Amount: count,
		})
	})
	return spell
}

// TestStormLineNamesTheCardAndTheCount is the line a player reads.
func TestStormLineNamesTheCardAndTheCount(t *testing.T) {
	g := buildActiveGame(t)
	spell := seedStormLog(t, g, 3)

	entry := findLog(t, ViewOfGame(g).Log, LogStorm)
	if entry.Amount != 3 {
		t.Errorf("amount = %d, want 3", entry.Amount)
	}
	if entry.CardID != spell.String() {
		t.Errorf("card_id = %q, want the storm spell", entry.CardID)
	}
	if entry.Seat != 0 {
		t.Errorf("seat = %d, want the caster's seat 0", entry.Seat)
	}
	if want := "Grapeshot — storm count 3"; entry.Text != want {
		t.Errorf("text = %q, want %q", entry.Text, want)
	}
}

// TestStormLineIsEmittedAtZero — "the trigger resolved and found
// nothing to copy" and "the trigger never fired" are different facts
// and the log is the only place the difference shows.
func TestStormLineIsEmittedAtZero(t *testing.T) {
	g := buildActiveGame(t)
	seedStormLog(t, g, 0)

	entry := findLog(t, ViewOfGame(g).Log, LogStorm)
	if entry.Amount != 0 {
		t.Errorf("amount = %d, want 0", entry.Amount)
	}
	if !strings.Contains(entry.Text, "storm count 0") {
		t.Errorf("text = %q, want it to say the count is zero", entry.Text)
	}
}

// TestStormCountSurvivesRedaction — the count is a fact about the
// turn's CASTS, which every seat watched, not a value read off the
// card. So unlike a counter total, a Saga chapter or a Class level,
// it does not travel with the name when the name is dropped.
func TestStormCountSurvivesRedaction(t *testing.T) {
	entry := LogEvent{Kind: LogStorm, Amount: 3}
	if entry.amountNamesTheCard() {
		t.Error("a storm count is redacted with the card name — it should not be")
	}
	if got := renderLogText(entry, "", ""); got != "a card — storm count 3" {
		t.Errorf("redacted text = %q, want %q", got, "a card — storm count 3")
	}
}
