package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const stGiantShark = "44a10a63-be9c-4f1d-aad6-b5337112bda5"

// Giant Shark: blocking a creature that was dealt damage this turn pumps
// it; blocking an undamaged one does not.
func TestGiantSharkPumpsAgainstADamagedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Island", "", "Basic Land — Island", 0, 0))
	shark := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Giant Shark", stGiantShark, "Creature — Shark", 4, 4))
	fresh := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	hurt := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.StCard("Wall", "", "Creature — Wall", 0, 5))
	block := func(attacker uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: shark, Target: attacker, Amount: 1})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	block(fresh)
	if c := effectstest.ApaLive(g, shark); c.CurrentPower() != 4 {
		t.Fatalf("blocking an undamaged creature: power %d, want 4", c.CurrentPower())
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatal(err)
		}
	})
	block(hurt)
	c := effectstest.ApaLive(g, shark)
	if c.CurrentPower() != 6 || !game.HasKeyword(c, "trample") {
		t.Fatalf("blocking a damaged creature: power %d trample %v, want 6 and trample", c.CurrentPower(), game.HasKeyword(c, "trample"))
	}
}

// Giant Shark is sacrificed when its controller's last Island goes.
func TestGiantSharkNeedsAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	island := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Island", "", "Basic Land — Island", 0, 0))
	shark := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Giant Shark", stGiantShark, "Creature — Shark", 4, 4))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, shark); c == nil || len(c.Effective().AttackTargetRestrictions) == 0 {
		t.Fatal("Giant Shark has no attack-target restriction")
	}
	effectstest.StDestroy(t, g, island)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, shark) {
		t.Fatal("Giant Shark survived the loss of its last Island")
	}
}
