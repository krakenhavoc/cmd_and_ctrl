package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// seedColoredCreature puts a 2/2 on the battlefield with the given
// colours (nil for colourless) and +1/+1 counters.
func SeedColoredCreature(g *game.Game, owner uuid.UUID, name string, colors []string, plusOnes int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: TestCreatureTypeLine,
		Power: 2, Toughness: 2, Colors: colors, Owner: owner, Controller: owner,
	}
	if plusOnes > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: plusOnes}
	}
	return PushBattlefieldCardWithTimestamp(g, c)
}

// strictEquip floats `mana` and activates `equipment`'s equip ability
// (its only activated ability) on `target` with the payment enforced.
// Returns the activation's error and leaves the item on the stack.
func StrictEquip(t *testing.T, g *game.Game, controller, equipment, target uuid.UUID, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(controller, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(controller, equipment, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		Strict:  true,
	})
}

func PoolSize(g *game.Game, player uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == player {
			return p.ManaPool.Total()
		}
	}
	return -1
}

func AttachedTo(t *testing.T, g *game.Game, equipment, host uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, equipment)
	if !ok {
		t.Fatalf("equipment %s left the battlefield", equipment)
	}
	return c.IsAttachedTo(host)
}

// --- helpers --------------------------------------------------------

// activatedRowOf reads `source`'s activated ability row `index` from
// the viewer's frame.
func ActivatedRowOf(t *testing.T, g *game.Game, viewer, source uuid.UUID, index int) protocol.ActivatedAbilityView {
	t.Helper()
	// The seeding helpers push straight onto the battlefield without
	// the entry path's knower stamp, and the per-viewer filter redacts
	// a card nobody is marked as knowing. A permanent is public.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, p := range g.Seats {
				c.KnownBy[p.ID] = true
			}
		}
	})
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != source.String() {
			continue
		}
		for _, row := range cv.ActivatedAbilities {
			if row.Index == index {
				return row
			}
		}
	}
	t.Fatalf("no activated ability row %d on %s", index, source)
	return protocol.ActivatedAbilityView{}
}
