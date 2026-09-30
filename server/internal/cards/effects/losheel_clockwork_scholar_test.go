package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const losheelOracle = "f8c2a972-e38f-47e5-a355-6f30ad09b1ae"

// TestLosheelPreventsCombatDamageToAttackingArtifactCreaturesYouControl
// — a blocker's combat damage to your own attacking artifact creature
// is prevented; damage to a non-artifact attacker is not.
func TestLosheelPreventsCombatDamageToAttackingArtifactCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Losheel, Clockwork Scholar", "Legendary Creature — Elephant Artificer", losheelOracle, false)

	attacker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: attacker, Name: "Attacking Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, AttackingTarget: opp.ID,
	})
	blocker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: blocker, Name: "Blocker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.MarkCombatDamage(blocker, attacker, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == attacker {
				marked = c.DamageMarked
			}
		}
	})
	if marked != 0 {
		t.Errorf("DamageMarked on the attacking artifact creature = %d, want 0", marked)
	}

	// A non-artifact attacker still takes its damage.
	nonArtifactAttacker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: nonArtifactAttacker, Name: "Plain Attacker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 5, Owner: me.ID, Controller: me.ID, AttackingTarget: opp.ID,
	})
	if err := g.MarkCombatDamage(blocker, nonArtifactAttacker, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == nonArtifactAttacker {
				marked = c.DamageMarked
			}
		}
	})
	if marked != 2 {
		t.Errorf("DamageMarked on a non-artifact attacker = %d, want 2 (Losheel should not prevent it)", marked)
	}
}

// TestLosheelDrawsOnceEachTurnWhenArtifactCreaturesEnter — a batch of
// artifact creatures entering together draws one card; a second
// separate entry the same turn draws none.
func TestLosheelDrawsOnceEachTurnWhenArtifactCreaturesEnter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Losheel, Clockwork Scholar", "Legendary Creature — Elephant Artificer", losheelOracle, false)

	handBefore := me.Hand.Size()
	firstID := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: firstID, Name: "Construct One", TypeLine: "Artifact Creature — Construct",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, firstID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Construct One: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand after first artifact creature enters: %d, want %d (cast one, drew one)", got, handBefore+1)
	}

	secondID := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: secondID, Name: "Construct Two", TypeLine: "Artifact Creature — Construct",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	handBefore2 := me.Hand.Size()
	if err := g.CastSpell(me.ID, secondID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Construct Two: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore2-1 {
		t.Errorf("hand after SECOND artifact creature enters this turn: %d, want %d (cast one, no second draw)", got, handBefore2-1)
	}
}
