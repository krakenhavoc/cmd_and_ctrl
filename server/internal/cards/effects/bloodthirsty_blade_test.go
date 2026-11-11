package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bloodthirstyBladeOracle = "b2747459-05ec-42a6-81be-e97e182bc778"

func TestBloodthirstyBladeIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(bloodthirstyBladeOracle)
	if !ok || spec.Name != "Bloodthirsty Blade" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
	for _, ab := range spec.Activated {
		if ab.Equip {
			t.Error("the {1} attach ability is not the equip keyword")
		}
	}
}

func TestBloodthirstyBladeAttachesToAnOpponentsCreatureAndGoadsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	blade := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bloodthirsty Blade", TypeLine: equipTypeLine,
		OracleID: bloodthirstyBladeOracle, Owner: me.ID, Controller: me.ID,
	})
	mine := tdCreature(g, me.ID, "Mine", 2, 2)
	theirs := tdCreature(g, opp.ID, "Theirs", 2, 2)

	// Only a creature an opponent controls is a legal target.
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}, Strict: true,
	}); err == nil {
		t.Fatal("the Blade attached to the controller's own creature")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, _ := g.LookupCardForEffect(theirs)
	if c.Controller != opp.ID {
		t.Fatalf("control changed: %v", c.Controller)
	}
	if p := c.PowerForComparison(); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
	if !c.Goaded() {
		t.Error("the equipped creature is not goaded")
	}

	// On its controller's turn it must attack, and not the Blade's controller.
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.Clone().DeclareAttacker(theirs, me.ID); err == nil {
		t.Error("the goaded creature attacked the goader while another player was open")
	}
	if err := g.Clone().DeclareAttacker(theirs, third.ID); err != nil {
		t.Errorf("attacking another player: %v", err)
	}
}
