package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_control_cards_test.go — the proof cards for #1317 (CR 724.2,
// end the combat phase) and #1329 (CR 508.7, reselect an attack).

const (
	mandateOfPeaceOracle     = "c309ec42-34a0-4083-a36c-7814643d7960"
	misleadingSignpostOracle = "eaffdf95-8e20-408d-99fa-3adc9e19523d"
	portalMageOracle         = "4d19ea6d-cabe-4a25-a911-dd744517dd73"
	windshaperOracle         = "c47050c2-9b46-427e-85d6-058fd6a61e67"
)

// --- Mandate of Peace (#1317) ---------------------------------------

// Cast in the declare blockers step on top of another spell: the
// stack is exiled (Mandate included), the attack is over, no combat
// damage is dealt, and the turn is in its postcombat main phase.
func TestMandateOfPeaceEndsTheCombatPhase(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, me.ID, "Charger", 5, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	under := castInPlace(t, g, me.ID, "Some Instant", "")
	mandate := castInPlace(t, g, opp.ID, "Mandate of Peace", mandateOfPeaceOracle)
	life := opp.Life
	passPriorityAroundTable(t, g)

	if g.Turn.Step != game.StepPostcombatMain {
		t.Fatalf("at %s, want postcombat_main", g.Turn.Step)
	}
	for _, id := range []uuid.UUID{mandate, under} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s is not in exile; the whole stack is exiled (CR 724.2b)", id)
		}
	}
	if c, ok := aangCardOnBF(g, attacker); !ok || c.AttackingTarget != uuid.Nil {
		t.Errorf("the attacker is still in combat")
	}
	if opp.Life != life {
		t.Errorf("the defender lost %d life; the combat ended before damage", life-opp.Life)
	}
}

// "Cast this spell only during combat."
func TestMandateOfPeaceCastOnlyDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Mandate of Peace", TypeLine: "Instant",
		OracleID: mandateOfPeaceOracle, Owner: me.ID, Controller: me.ID,
	})
	var cantCast *game.CantCastError
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); !errors.As(err, &cantCast) {
		t.Errorf("cast in the precombat main phase: err = %v, want a CantCastError", err)
	}
}

// --- Reselecting an attack (#1329) ------------------------------------

// castFlashPermanentInPlace casts a flash permanent without moving the
// turn on, so the declare attackers step is still where it resolves.
func castFlashPermanentInPlace(t *testing.T, g *game.Game, caster uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	var p *game.Player
	for _, s := range g.Seats {
		if s != nil && s.ID == caster {
			p = s
		}
	}
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		OracleID: oracle, Owner: caster, Controller: caster,
	})
	if err := g.CastSpell(caster, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// attackInDeclareAttackers declares the attacks and leaves the cursor
// in the declare attackers step.
func attackInDeclareAttackers(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range attacks {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
}

// settleUntilOptionPick passes priority, answering a pick_target with
// `target`, until an option pick is owed by `chooser` (returned) or
// the stack is empty (nil).
func settleUntilOptionPick(t *testing.T, g *game.Game, chooser, target uuid.UUID) *game.PendingChoice {
	t.Helper()
	for range 8 {
		if c := latestOptionPickFor(g, chooser); c != nil {
			return c
		}
		if pick := latestPickTarget(g, chooser); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return nil
		}
		passPriorityAroundTable(t, g)
	}
	return latestOptionPickFor(g, chooser)
}

// optionFor is the index of the option naming `subject` as a seat.
func optionFor(t *testing.T, c *game.PendingChoice, seat uuid.UUID) int {
	t.Helper()
	for i, opt := range c.PickOptions {
		if opt.Player == seat {
			return i
		}
	}
	t.Fatalf("no option for seat %s in %+v", seat, c.PickOptions)
	return -1
}

// The deck card: the defending player flashes in Misleading Signpost
// and pushes the attack onto another opponent of the attacker.
func TestMisleadingSignpostRedirectsAnAttackAimedAtYou(t *testing.T) {
	g := newCatalogGame(t)
	attackerSeat, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
	attacker := pushVanillaCreature(g, attackerSeat.ID, "Charger", 4, 4)
	attackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{attacker: me.ID})

	castFlashPermanentInPlace(t, g, me.ID, "Misleading Signpost", "Artifact", misleadingSignpostOracle)
	c := settleUntilOptionPick(t, g, me.ID, attacker)
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
	if err := g.ResolveOptionPick(c.ID, me.ID, optionFor(t, c, other.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	card, _ := aangCardOnBF(g, attacker)
	if card.AttackingTarget != other.ID {
		t.Fatalf("the attack is on %v, want seat 2", card.AttackingTarget)
	}

	lifeMe, lifeOther := me.Life, other.Life
	advanceTo(t, g, game.StepPostcombatMain)
	if me.Life != lifeMe || other.Life != lifeOther-4 {
		t.Errorf("damage: me %d→%d, other %d→%d; the 4 should hit the reselected player",
			lifeMe, me.Life, lifeOther, other.Life)
	}
}

// Outside the declare attackers step the Signpost enters with no
// trigger — the condition is part of the trigger event.
func TestMisleadingSignpostOutsideDeclareAttackersDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	attackerSeat, me := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, attackerSeat.ID, "Charger", 4, 4)
	attackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{attacker: me.ID})
	advanceTo(t, g, game.StepDeclareBlockers)

	castFlashPermanentInPlace(t, g, me.ID, "Misleading Signpost", "Artifact", misleadingSignpostOracle)
	if c := settleUntilOptionPick(t, g, me.ID, attacker); c != nil {
		t.Errorf("a reselect prompt opened in the declare blockers step")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt opened outside the declare attackers step: %+v", g.PendingChoices[0])
	}
}

// Portal Mage is the same trigger on a body; "keep" leaves the attack
// where it was.
func TestPortalMageKeepLeavesTheAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	attackerSeat, me := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, attackerSeat.ID, "Charger", 4, 4)
	attackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{attacker: me.ID})

	castFlashPermanentInPlace(t, g, me.ID, "Portal Mage", "Creature — Human Wizard", portalMageOracle)
	c := settleUntilOptionPick(t, g, me.ID, attacker)
	if c == nil {
		t.Fatalf("no reselect prompt")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if card, _ := aangCardOnBF(g, attacker); card.AttackingTarget != me.ID {
		t.Errorf("\"keep\" moved the attack to %v", card.AttackingTarget)
	}
}

// Windshaper Planetar asks once per attacking creature, one after the
// other, and each answer is its own.
func TestWindshaperPlanetarAsksForEachAttacker(t *testing.T) {
	g := newCatalogGame(t)
	attackerSeat, me, second, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := pushVanillaCreature(g, attackerSeat.ID, "Charger A", 2, 2)
	b := pushVanillaCreature(g, attackerSeat.ID, "Charger B", 2, 2)
	attackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: me.ID, b: me.ID})

	castFlashPermanentInPlace(t, g, me.ID, "Windshaper Planetar", "Creature — Angel", windshaperOracle)
	first := settleUntilOptionPick(t, g, me.ID, uuid.Nil)
	if first == nil {
		t.Fatalf("no first reselect prompt")
	}
	if err := g.ResolveOptionPick(first.ID, me.ID, optionFor(t, first, second.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	next := latestOptionPickFor(g, me.ID)
	if next == nil || next.ID == first.ID {
		t.Fatalf("no second prompt for the second attacker")
	}
	if err := g.ResolveOptionPick(next.ID, me.ID, optionFor(t, next, third.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Errorf("a third prompt opened for two attackers")
	}
	ca, _ := aangCardOnBF(g, a)
	cb, _ := aangCardOnBF(g, b)
	got := map[uuid.UUID]bool{ca.AttackingTarget: true, cb.AttackingTarget: true}
	if !got[second.ID] || !got[third.ID] {
		t.Errorf("attacks ended on %v and %v, want seats 2 and 3 one each", ca.AttackingTarget, cb.AttackingTarget)
	}
}
