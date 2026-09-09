package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const aetherizeOracle = "7c779721-cd1b-4696-9ae9-68ccc284ed2a"

// Attackers go back to hand; a creature that stayed home does not.
func TestAetherizeReturnsOnlyAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	attacker := pushVanillaCreature(g, me.ID, "Charger", 3, 3)
	homebody := pushVanillaCreature(g, me.ID, "Homebody", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	// NOT castCatalogSpell: that helper advances to a main phase,
	// which walks past end of combat and clears AttackingTarget — so
	// any card reading combat state has to be cast in place.
	castInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	passPriorityAroundTable(t, g)

	if _, stillOut := aangCardOnBF(g, attacker); stillOut {
		t.Error("attacking creature was not returned to hand")
	}
	if _, stillOut := aangCardOnBF(g, homebody); !stillOut {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	// Seat 1 attacks seat 2 on seat 1's turn; seat 0 flashes it in.
	oppAttacker := pushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	aangAdvanceToMain(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(oppAttacker, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	castInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	passPriorityAroundTable(t, g)

	if _, stillOut := aangCardOnBF(g, oppAttacker); stillOut {
		t.Error("an opponent's attacker was not returned")
	}
}

// Nothing attacking is a legal, harmless resolution.
func TestAetherizeWithNoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	castInPlace(t, g, me.ID, "Aetherize", aetherizeOracle)
	passPriorityAroundTable(t, g)

	if _, stillOut := aangCardOnBF(g, bear); !stillOut {
		t.Error("Aetherize bounced a creature with no combat declared")
	}
}

// castInPlace casts an instant from caster's hand WITHOUT advancing
// the turn, so the current step (and any combat declarations) survive
// into resolution.
func castInPlace(t *testing.T, g *game.Game, caster uuid.UUID, name, oracle string) uuid.UUID {
	t.Helper()
	var p *game.Player
	for _, s := range g.Seats {
		if s != nil && s.ID == caster {
			p = s
		}
	}
	if p == nil {
		t.Fatalf("no such seat %s", caster)
	}
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		OracleID: oracle, Owner: caster, Controller: caster,
	})
	if err := g.CastSpell(caster, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
