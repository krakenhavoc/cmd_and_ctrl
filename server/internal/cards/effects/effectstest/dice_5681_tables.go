package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// d5681Table is a table at seat 0's main phase whose next d20 from
// `source` shows `natural`.
func D5681Table(t *testing.T, source uuid.UUID, natural int) *game.Game {
	t.Helper()
	key := D5681Key(t, source, 20, 1, func(r []int) bool { return r[0] == natural })
	g := NewCatalogGame(t)
	g.SetRNGKeyForTest(key)
	AdvanceTo(t, g, game.StepPrecombatMain)
	return g
}

// d5681Spell puts a spell with a chosen instance ID in the player's hand.
func D5681Spell(g *game.Game, p *game.Player, id uuid.UUID, name, typeLine, cost, oracle string) {
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID,
		})
	})
}

// d5681Dies moves a creature from the battlefield to its owner's
// graveyard.
func D5681Dies(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("to the graveyard: %v", err)
	}
}
