package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Mandate of Peace (#1317) ---------------------------------------

// Cast in the declare blockers step on top of another spell: the
// stack is exiled (Mandate included), the attack is over, no combat
// damage is dealt, and the turn is in its postcombat main phase.
func TestMandateOfPeaceEndsTheCombatPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Charger", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)

	under := effectstest.CastInPlace(t, g, me.ID, "Some Instant", "")
	mandate := effectstest.CastInPlace(t, g, opp.ID, "Mandate of Peace", effectstest.MandateOfPeaceOracle)
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)

	if g.Turn.Step != game.StepPostcombatMain {
		t.Fatalf("at %s, want postcombat_main", g.Turn.Step)
	}
	for _, id := range []uuid.UUID{mandate, under} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s is not in exile; the whole stack is exiled (CR 724.2b)", id)
		}
	}
	if c, ok := effectstest.AangCardOnBF(g, attacker); !ok || c.AttackingTarget != uuid.Nil {
		t.Errorf("the attacker is still in combat")
	}
	if opp.Life != life {
		t.Errorf("the defender lost %d life; the combat ended before damage", life-opp.Life)
	}
}

// "Cast this spell only during combat."
func TestMandateOfPeaceCastOnlyDuringCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Mandate of Peace", TypeLine: "Instant",
		OracleID: effectstest.MandateOfPeaceOracle, Owner: me.ID, Controller: me.ID,
	})
	var cantCast *game.CantCastError
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); !errors.As(err, &cantCast) {
		t.Errorf("cast in the precombat main phase: err = %v, want a CantCastError", err)
	}
}

// The deck card: the defending player flashes in Misleading Signpost
// and pushes the attack onto another opponent of the attacker.
func TestMisleadingSignpostRedirectsAnAttackAimedAtYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
	attacker := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger", 4, 4)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{attacker: me.ID})

	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Misleading Signpost", "Artifact", effectstest.MisleadingSignpostOracle)
	c := effectstest.SettleUntilOptionPick(t, g, me.ID, attacker)
	if c == nil {
		t.Fatalf("no reselect prompt reached the Signpost's controller")
	}
	if c.PickOptions[0].Player != uuid.Nil {
		t.Fatalf("the first option is not \"keep\": %+v", c.PickOptions[0])
	}
	for _, opt := range c.PickOptions {
		if opt.Player == attackerSeat.ID {
			t.Errorf("the attacker's own controller is offered (CR 508.7c)")
		}
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, effectstest.OptionFor(t, c, other.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	card, _ := effectstest.AangCardOnBF(g, attacker)
	if card.AttackingTarget != other.ID {
		t.Fatalf("the attack is on %v, want seat 2", card.AttackingTarget)
	}

	lifeMe, lifeOther := me.Life, other.Life
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if me.Life != lifeMe || other.Life != lifeOther-4 {
		t.Errorf("damage: me %d→%d, other %d→%d; the 4 should hit the reselected player",
			lifeMe, me.Life, lifeOther, other.Life)
	}
}

// Outside the declare attackers step the Signpost enters with no
// trigger — the condition is part of the trigger event.
func TestMisleadingSignpostOutsideDeclareAttackersDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger", 4, 4)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{attacker: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)

	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Misleading Signpost", "Artifact", effectstest.MisleadingSignpostOracle)
	if c := effectstest.SettleUntilOptionPick(t, g, me.ID, attacker); c != nil {
		t.Errorf("a reselect prompt opened in the declare blockers step")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt opened outside the declare attackers step: %+v", g.PendingChoices[0])
	}
}

// Portal Mage is the same trigger on a body; "keep" leaves the attack
// where it was.
func TestPortalMageKeepLeavesTheAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger", 4, 4)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{attacker: me.ID})

	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Portal Mage", "Creature — Human Wizard", effectstest.PortalMageOracle)
	c := effectstest.SettleUntilOptionPick(t, g, me.ID, attacker)
	if c == nil {
		t.Fatalf("no reselect prompt")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if card, _ := effectstest.AangCardOnBF(g, attacker); card.AttackingTarget != me.ID {
		t.Errorf("\"keep\" moved the attack to %v", card.AttackingTarget)
	}
}

// Windshaper Planetar asks once per attacking creature, one after the
// other, and each answer is its own.
func TestWindshaperPlanetarAsksForEachAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me, second, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger A", 2, 2)
	b := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger B", 2, 2)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: me.ID, b: me.ID})

	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Windshaper Planetar", "Creature — Angel", effectstest.WindshaperOracle)
	first := effectstest.SettleUntilOptionPick(t, g, me.ID, uuid.Nil)
	if first == nil {
		t.Fatalf("no first reselect prompt")
	}
	if err := g.ResolveOptionPick(first.ID, me.ID, effectstest.OptionFor(t, first, second.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	next := effectstest.LatestOptionPickFor(g, me.ID)
	if next == nil || next.ID == first.ID {
		t.Fatalf("no second prompt for the second attacker")
	}
	if err := g.ResolveOptionPick(next.ID, me.ID, effectstest.OptionFor(t, next, third.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Errorf("a third prompt opened for two attackers")
	}
	ca, _ := effectstest.AangCardOnBF(g, a)
	cb, _ := effectstest.AangCardOnBF(g, b)
	got := map[uuid.UUID]bool{ca.AttackingTarget: true, cb.AttackingTarget: true}
	if !got[second.ID] || !got[third.ID] {
		t.Errorf("attacks ended on %v and %v, want seats 2 and 3 one each", ca.AttackingTarget, cb.AttackingTarget)
	}
}
