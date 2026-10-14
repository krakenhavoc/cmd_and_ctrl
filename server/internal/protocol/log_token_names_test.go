package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_token_names_test.go — #3609. A log line that mentions a token
// keeps the token's name after the token has ceased to exist, for every
// viewer, and the fallback never names anything hidden.

// tokenOnBattlefield seeds a face-up token every seat knows.
func tokenOnBattlefield(t *testing.T, g *game.Game, name string, controller uuid.UUID) uuid.UUID {
	t.Helper()
	return battlefieldCard(t, g, name, "Token Creature — Goblin", controller)
}

// logTexts is every viewer's rendering of the log, so a leak or a lost
// name on any seat or the spectator fails the test.
func logTextsFor(g *game.Game) map[string][]string {
	out := map[string][]string{}
	for _, viewer := range []string{g.Seats[0].ID.String(), g.Seats[1].ID.String(), SpectatorViewerID} {
		for _, e := range FilterViewFor(ViewOfGame(g), viewer).Log {
			out[viewer] = append(out[viewer], e.Text)
		}
	}
	return out
}

// The three lines of the report that a token produces: a counter put on
// it, a sacrifice of it, and the move that says it ceased to exist.
func TestLogKeepsATokensNameAfterItCeasesToExist(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	goblin := tokenOnBattlefield(t, g, "Goblin Instigator", me.ID)
	treasure := tokenOnBattlefield(t, g, "Treasure", me.ID)

	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(goblin, "slime", 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
		if err := g.SacrificePermanentForEffect(treasure); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(goblin); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	// Both are now in the graveyard; the sweep removes them.
	g.RunStateChecksForTest()

	for _, id := range []uuid.UUID{goblin, treasure} {
		if _, ok := g.LookupCardForEffect(id); ok {
			t.Fatalf("token %s still exists; the test setup is wrong", id)
		}
	}
	for viewer, texts := range logTextsFor(g) {
		joined := strings.Join(texts, "\n")
		for _, want := range []string{
			"Goblin Instigator now has 1 slime counter",
			"sacrificed Treasure",
			"Treasure ceased to exist",
		} {
			if !strings.Contains(joined, want) {
				t.Errorf("viewer %s: log lacks %q:\n%s", viewer, want, joined)
			}
		}
		if strings.Contains(joined, "moved from") && strings.Contains(joined, "to elsewhere") {
			t.Errorf("viewer %s: a token ceasing to exist still reads as a move:\n%s", viewer, joined)
		}
	}
}

// The spell path of a counter stamps no Source, so the line read "a card
// countered Yuna" even with the counterspell sitting in a graveyard. The
// resolving spell is its cause item, and its card is what the line names.
func TestLogNamesTheSpellThatCountered(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	counter, victim := uuid.New(), uuid.New()
	known := map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.WithWriteLock(func() {
		g.Seats[0].Graveyard.PushTop(game.Card{
			InstanceID: counter, Name: "Absorb", TypeLine: "Instant", Owner: me.ID, KnownBy: known,
		})
		g.Seats[1].Graveyard.PushTop(game.Card{
			InstanceID: victim, Name: "Yuna, Grand Summoner", TypeLine: "Creature", Owner: opp.ID, KnownBy: known,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCounterSpell, Actor: me.ID, CardID: victim, Target: victim,
			Cause: game.MoveCauseEffect, CauseController: me.ID, CauseItem: counter,
		})
	})
	entry := findLog(t, FilterViewFor(ViewOfGame(g), opp.ID.String()).Log, LogCounter)
	if want := "Absorb countered Yuna, Grand Summoner"; entry.Text != want {
		t.Errorf("line = %q, want %q", entry.Text, want)
	}
}

// The fallback is exactly the captured name: a hidden card the view can
// no longer account for still reads "a card", and an event that carried
// no public name does not acquire one.
func TestLogFallbackDoesNotNameAHiddenCard(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	secret := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[1].Hand.PushTop(game.Card{
			InstanceID: secret, Name: "Secret Tutor", TypeLine: "Sorcery",
			Owner: opp.ID, Controller: opp.ID,
			KnownBy: map[uuid.UUID]bool{opp.ID: true},
		})
		g.EmitEvent(game.Event{Kind: game.EventTapCard, Actor: opp.ID, CardID: secret})
	})
	// The card is gone from every zone by the time the log renders.
	g.WithWriteLock(func() {
		if _, err := g.Seats[1].Hand.Remove(secret); err != nil {
			t.Fatalf("remove: %v", err)
		}
	})

	for _, viewer := range []string{me.ID.String(), SpectatorViewerID} {
		for _, e := range FilterViewFor(ViewOfGame(g), viewer).Log {
			if strings.Contains(e.Text, "Secret Tutor") {
				t.Errorf("viewer %s: leaked a hidden card's name: %q", viewer, e.Text)
			}
		}
	}
}

// A captured name belongs to a token every seat knew. A viewer the
// token was never public to cannot be handed it by the fallback: here
// the event carries a name for an object the view still finds and that
// the viewer may not identify, so the view's knowers win.
func TestLogCapturedNameYieldsToTheViewsKnowers(t *testing.T) {
	g := buildActiveGame(t)
	opp := g.Seats[1]
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[1].Hand.PushTop(game.Card{
			InstanceID: id, Name: "Real Name", TypeLine: "Instant",
			Owner: opp.ID, KnownBy: map[uuid.UUID]bool{opp.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventTapCard, Actor: opp.ID, CardID: id,
			PublicNames: map[uuid.UUID]string{id: "Forged Name"},
		})
	})
	for _, e := range FilterViewFor(ViewOfGame(g), g.Seats[0].ID.String()).Log {
		if strings.Contains(e.Text, "Forged Name") || strings.Contains(e.Text, "Real Name") {
			t.Errorf("a hand card's line names it for a non-knower: %q", e.Text)
		}
	}
}
