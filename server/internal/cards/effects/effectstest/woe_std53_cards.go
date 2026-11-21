package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Dream Spoilers --------------------------------------------------------------

func Std53CastEvent(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	spell := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: spell, Name: "Some Instant", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: p.ID, CardID: spell})
	})
}
