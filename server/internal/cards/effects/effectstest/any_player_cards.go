package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// apaPush puts `c` on the battlefield owned by `owner` and controlled
// by `controller`, as an entry the engine has seen, free of summoning
// sickness, with the layers caught up.
func ApaPush(g *game.Game, owner, controller uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, controller
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: c.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		FindBattlefieldCardForTest(g, c.InstanceID).SummonedThisTurn = false
		g.RecomputeLayersIfStaleLocked()
	})
	return c.InstanceID
}

// apaCreature is a creature card template.
func ApaCreature(name, oracle string, power, toughness int) game.Card {
	return game.Card{Name: name, OracleID: oracle, TypeLine: "Creature — Test", Power: power, Toughness: toughness}
}

// apaMana adds one mana of each listed color to p's pool.
func ApaMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// apaLive is the permanent with the layers caught up.
func ApaLive(g *game.Game, id uuid.UUID) *game.Card {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return FindBattlefieldCardForTest(g, id)
}

// apaActivate has `p` activate row `index` of `source` and resolves it.
func ApaActivate(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	params.Strict = true
	if err := g.ActivateCatalogAbility(p.ID, source, index, params); err != nil {
		t.Fatalf("%s activates row %d: %v", p.Name, index, err)
	}
	if n := p.ManaPool.Total(); n != 0 {
		t.Errorf("%s's pool holds %d after paying, want 0", p.Name, n)
	}
	PassPriorityAroundTable(t, g)
}
