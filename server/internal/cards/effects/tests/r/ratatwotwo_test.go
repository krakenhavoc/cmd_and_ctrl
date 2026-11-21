package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRatatwotwoEntersAndGivesAChefRoleToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rat := effectstest.CastCatalogSpell(t, g, "Ratatwotwo", "Legendary Creature — Rat", effectstest.RatatwotwoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if roles := effectstest.RolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Chef Role" {
		t.Fatalf("roles on the bear = %+v, want a Chef Role", roles)
	}
	if roles := effectstest.RolesOn(g, rat); len(roles) != 0 {
		t.Error("Ratatwotwo put the Role on itself")
	}
}

func TestRatatwotwoPumpsChefEnchantedAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushRatatwotwo(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleChef, bear, me.ID)

	effectstest.DeclareAttack(t, g, opp.ID, bear, other)
	// Two triggers at once from the enchanted creature's attack: the
	// Chef Role's Food and Ratatwotwo's pump (CR 603.3b order prompt).
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// 2/2 +1/+1 (Chef Role) +2/+2 (Ratatwotwo's power) until end of
	// turn; only the enchanted one.
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("enchanted attacker has power %d, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, other); p != 2 {
		t.Errorf("unenchanted attacker has power %d, want 2", p)
	}
	if effectstest.CountOnBattlefield(g, "Food", me.ID) != 1 {
		t.Errorf("Food tokens = %d, want 1 (the Chef Role's own trigger)", effectstest.CountOnBattlefield(g, "Food", me.ID))
	}
}

func TestRatatwotwoPumpsAChefEnchantedBlockerOncePerBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.PushRatatwotwo(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.GiveRole(t, g, effects.RoleChef, bear, me.ID)
	attacker := effectstest.PushVanillaCreature(g, opp.ID, "Attacker", 2, 2)

	for i := 1; i <= 2; i++ {
		n := i
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, CardID: bear, Target: attacker, Actor: me.ID, Amount: n})
		})
	}
	if n := effectstest.TriggersOnStackFrom(g, rat) + len(g.PendingTriggers); n != 1 {
		t.Fatalf("Ratatwotwo triggers for one blocker blocking two attackers = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("enchanted blocker has power %d, want 5", p)
	}
}
