package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ratatwotwoOracle = "0edd7d9a-af37-4d3b-8b4b-913c36dadf82"

// pushRatatwotwo seeds the real 2/2 on the battlefield.
func pushRatatwotwo(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ratatwotwo", OracleID: ratatwotwoOracle,
		TypeLine: "Legendary Creature — Rat", Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func TestRatatwotwoEntersAndGivesAChefRoleToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rat := castCatalogSpell(t, g, "Ratatwotwo", "Legendary Creature — Rat", ratatwotwoOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Chef Role" {
		t.Fatalf("roles on the bear = %+v, want a Chef Role", roles)
	}
	if roles := rolesOn(g, rat); len(roles) != 0 {
		t.Error("Ratatwotwo put the Role on itself")
	}
}

func TestRatatwotwoPumpsChefEnchantedAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushRatatwotwo(g, me.ID)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other", 2, 2)
	giveRole(t, g, RoleChef, bear, me.ID)

	declareAttack(t, g, opp.ID, bear, other)
	// Two triggers at once from the enchanted creature's attack: the
	// Chef Role's Food and Ratatwotwo's pump (CR 603.3b order prompt).
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	// 2/2 +1/+1 (Chef Role) +2/+2 (Ratatwotwo's power) until end of
	// turn; only the enchanted one.
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("enchanted attacker has power %d, want 5", p)
	}
	if p := effectivePower(t, g, other); p != 2 {
		t.Errorf("unenchanted attacker has power %d, want 2", p)
	}
	if countOnBattlefield(g, "Food", me.ID) != 1 {
		t.Errorf("Food tokens = %d, want 1 (the Chef Role's own trigger)", countOnBattlefield(g, "Food", me.ID))
	}
}

func TestRatatwotwoPumpsAChefEnchantedBlockerOncePerBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := pushRatatwotwo(g, me.ID)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	giveRole(t, g, RoleChef, bear, me.ID)
	attacker := pushVanillaCreature(g, opp.ID, "Attacker", 2, 2)

	for i := 1; i <= 2; i++ {
		n := i
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, CardID: bear, Target: attacker, Actor: me.ID, Amount: n})
		})
	}
	if n := triggersOnStackFrom(g, rat) + len(g.PendingTriggers); n != 1 {
		t.Fatalf("Ratatwotwo triggers for one blocker blocking two attackers = %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("enchanted blocker has power %d, want 5", p)
	}
}
