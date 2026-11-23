package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- helpers ---------------------------------------------------------------

// s82b10Grave puts a card in a player's graveyard.
func S82b10Grave(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// s82b10Settle answers a trigger's target prompt, when one was raised, and
// lets the table settle.
func S82b10Settle(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.TargetRef) {
	t.Helper()
	if pick != nil {
		if p := LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, *pick); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
	}
	PassPriorityAroundTable(t, g)
}

func S82b10Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
}

func S82b10Target(id uuid.UUID) *game.TargetRef {
	r := EoeCDCard(id)
	return &r
}
