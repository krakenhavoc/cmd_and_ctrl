package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestCompoundWardMovesRequireAndPriceWholeCost(t *testing.T) {
	for _, tc := range []struct {
		life        int
		cost        string
		phy, locked bool
	}{
		{40, "{2}", false, false}, {1, "{2}", false, false}, {40, "{2}", false, true},
		{4, "{B/P}", true, false}, {3, "{B/P}", true, false},
	} {
		life := tc.life
		g := newTable(t)
		payer := g.Seats[1]
		payer.Life = life
		payer.ManaPool.AddMana(game.ManaToken{Color: "C"})
		payer.ManaPool.AddMana(game.ManaToken{Color: "C"})
		if tc.locked {
			g.WithWriteLock(func() { g.GrantLifeTotalLockForEffect(payer.ID, "Life lock", uuid.Nil, game.IndefiniteDuration()) })
		}
		spell := uuid.New()
		g.Stack.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Instant", Owner: payer.ID, Controller: payer.ID})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[spell] = &game.StackItem{ID: spell, Kind: game.StackItemSpell, Controller: payer.ID, Owner: payer.ID, SourceCardID: spell}
		g.WithWriteLock(func() {
			if err := g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{StackItem: spell, Cost: tc.cost, LifeCost: 2}); err != nil {
				t.Fatal(err)
			}
		})
		moves := legal.EnumerateFor(g, payer.ID)
		pays := 0
		for _, m := range moves {
			var p struct {
				Apply *bool `json:"apply"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if p.Apply != nil && *p.Apply {
				pays++
				if m.Cost == nil || m.Cost.Life != map[bool]int{false: 2, true: 4}[tc.phy] {
					t.Fatal("pay move omits mandatory life price")
				}
			}
		}
		if (pays > 0) != (!tc.locked && life >= map[bool]int{false: 2, true: 4}[tc.phy]) {
			t.Fatalf("life %d: %d pay moves", life, pays)
		}
		dispatchAll(t, g, payer.ID, moves)
	}
}
