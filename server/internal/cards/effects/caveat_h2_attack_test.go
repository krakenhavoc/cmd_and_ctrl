package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3172 cav-h2: The Aetherspark "can't be attacked" while attached to a
// creature (Characteristic.CantBeAttacked).
func TestAethersparkCantBeAttackedOnlyWhileAttached(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spark := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Aetherspark",
		TypeLine: "Legendary Artifact Planeswalker — Equipment", OracleID: theAethersparkOracle,
		Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	host := b12Creature(g, opp.ID, "Holder", "Creature — Bear", 2, 2)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Goblin", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	targets := func() []game.AttackTargetRef {
		var out []game.AttackTargetRef
		g.ReadSnapshot(func() { out = g.AttackTargetsForEffect(me.ID) })
		return out
	}
	offers := func(id uuid.UUID) bool {
		for _, tgt := range targets() {
			if tgt.ID == id {
				return true
			}
		}
		return false
	}
	if !offers(spark) {
		t.Fatal("an unattached Aetherspark is attackable like any planeswalker")
	}
	brAttach(t, g, spark, host)
	if offers(spark) {
		t.Error("an attached Aetherspark is offered as an attack target")
	}
	if err := g.DeclareAttacker(raider, spark); err == nil {
		t.Error("attacking an attached Aetherspark was accepted")
	}
	if err := g.DeclareAttacker(raider, opp.ID); err != nil {
		t.Errorf("the player behind it is still attackable: %v", err)
	}
}
