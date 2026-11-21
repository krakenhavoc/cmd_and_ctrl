package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveCreature puts a creature card with a real mana cost into p's
// graveyard — the mana value is the point of every test below.
func GraveCreature(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// castX casts a catalog spell from the active seat's hand announcing
// x, and returns the announce error rather than failing on it.
func CastX(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x, Targets: targets})
}
