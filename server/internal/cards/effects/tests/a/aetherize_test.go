package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const aetherizeOracle = "7c779721-cd1b-4696-9ae9-68ccc284ed2a"

// Attackers go back to hand; a creature that stayed home does not.
func TestAetherizeReturnsOnlyAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := effectstest.PushVanillaCreature(g, me.ID, "Charger", 3, 3)
	homebody := effectstest.PushVanillaCreature(g, me.ID, "Homebody", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	// NOT castCatalogSpell: that helper advances to a main phase,
	// which walks past end of combat and clears AttackingTarget — so
	// any card reading combat state has to be cast in place.
	effectstest.CastInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	effectstest.PassPriorityAroundTable(t, g)

	if _, stillOut := effectstest.AangCardOnBF(g, attacker); stillOut {
		t.Error("attacking creature was not returned to hand")
	}
	if _, stillOut := effectstest.AangCardOnBF(g, homebody); !stillOut {
		t.Error("non-attacking creature was returned; Aetherize hits attackers only")
	}
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == attacker {
			inHand = true
		}
	}
	if !inHand {
		t.Error("attacker did not end up in its owner's hand")
	}
}

// It is symmetric — it does not care whose attackers they are.
func TestAetherizeIsSymmetric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	// Seat 1 attacks seat 2 on seat 1's turn; seat 0 flashes it in.
	oppAttacker := effectstest.PushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(oppAttacker, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	effectstest.CastInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	effectstest.PassPriorityAroundTable(t, g)

	if _, stillOut := effectstest.AangCardOnBF(g, oppAttacker); stillOut {
		t.Error("an opponent's attacker was not returned")
	}
}

// The end of combat step is the last window a flash deck has, and
// creatures are in combat for all of it (CR 511.3, #785). Cast there,
// Aetherize still finds the attack.
func TestAetherizeCastInTheEndOfCombatStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := effectstest.PushVanillaCreature(g, me.ID, "Charger", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)

	effectstest.CastInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	effectstest.PassPriorityAroundTable(t, g)

	if _, stillOut := effectstest.AangCardOnBF(g, attacker); stillOut {
		t.Error("an attacker was not returned by Aetherize cast in the end of combat step")
	}
}

// Nothing attacking is a legal, harmless resolution.
func TestAetherizeWithNoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.CastInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	effectstest.PassPriorityAroundTable(t, g)

	if _, stillOut := effectstest.AangCardOnBF(g, bear); !stillOut {
		t.Error("Aetherize bounced a creature with no combat declared")
	}
}
