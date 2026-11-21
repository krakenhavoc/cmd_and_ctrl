package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures ------------------------------------------------------------------

// woeS2CastInstant has p cast a vanilla instant with the given cost and
// settles it.
func WoeS2CastInstant(t *testing.T, g *game.Game, p *game.Player, cost string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Some Instant", TypeLine: "Instant", ManaCost: cost, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a %s instant: %v", cost, err)
	}
	PassPriorityAroundTable(t, g)
}

// woeS2Block announces a block of attacker by blocker, as the lock-in does.
func WoeS2Block(g *game.Game, actor, blocker, attacker uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: actor, CardID: blocker, Target: attacker, Amount: 1})
	})
}
