package protocol

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3613: a resolution that paused on its own prompt writes its
// "resolved" line from EventResolutionFinished, after the answer, and
// none from its LogDeferred EventResolve.
func TestAPausedResolutionWritesItsResolvedLineWhenItFinishes(t *testing.T) {
	g := buildActiveGame(t)
	caster := g.Seats[0]
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spell, Name: "Council's Judgment", TypeLine: "Sorcery",
			Owner: caster.ID, Controller: caster.ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: caster.ID, Source: spell, CardID: spell,
			OldZone: game.ZoneHand, NewZone: game.ZoneStack})
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: caster.ID, Source: spell, CardID: spell, LogDeferred: true})
		g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: caster.ID})
		g.EmitEvent(game.Event{Kind: game.EventResolutionFinished, Actor: caster.ID, Source: spell, CardID: spell})
	})
	log := ViewOfGame(g).Log
	kinds := logKinds(log)
	resolves := 0
	for i, k := range kinds {
		if k != LogResolve {
			continue
		}
		resolves++
		if i == 0 || kinds[i-1] != LogDraw {
			t.Errorf("the resolved line is not after the effect's lines: %v", kinds)
		}
	}
	if resolves != 1 {
		t.Fatalf("%d resolved lines, want one: %v", resolves, kinds)
	}
	if got := findLog(t, log, LogResolve).Text; got != "Council's Judgment resolved" {
		t.Errorf("resolve text: got %q", got)
	}
}
