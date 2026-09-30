package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_phase_cards_test.go — the extra-combat half of ADR 0059's first
// wave (sub-PR 2b, #753). Each card is driven through the real cast,
// activation or attack, and the added phase is asserted by where the
// cursor goes next, never by reading the plan alone.

const (
	relentlessAssaultOracle     = "dc1c0e8c-d0c1-445c-968f-7dec91e5d5fc"
	seizeTheDayOracle           = "2a6abd59-e448-46f1-9af8-bb9040645971"
	aggravatedAssaultOracle     = "20129459-a386-41eb-899d-1aede3427300"
	fullThrottleOracle          = "3c2f87eb-b0f4-4578-b143-6fba313a82ec"
	karlachOracle               = "037355be-71e7-4866-80a6-80352c304970"
	hellkiteChargerOracle       = "d4f28a4b-d821-4132-bdec-4c528318f8e2"
	aureliaWarleaderOracle      = "0f5a3a09-2f07-4774-9e0f-e99d9a444166"
	sphinxOfTheSecondSunOracle  = "516101be-be39-4d84-8fee-d8a79930dd0a"
	eomerMarshalOfRohanOracle   = "b2d95950-18b3-463f-94f4-299e420751dc"
	extraPhaseTestCreatureType  = "Creature — Test"
	extraPhaseTestLegendaryType = "Legendary Creature — Test"
)

// stepOnce advances one step and returns where the cursor landed.
func stepOnce(t *testing.T, g *game.Game) game.Turn {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	return g.Turn
}

// countResolved counts EventResolve entries with `label` since `from`.
func countResolved(g *game.Game, from int, label string) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventResolve && ev.Label == label {
			n++
		}
	}
	return n
}

func TestRelentlessAssaultUntapsAttackersAndAddsCombatAndMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	idle := pushVanillaCreature(g, me.ID, "Idle", 2, 2)
	attackWith(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepPostcombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(idle) })
	if !isTapped(g, bear) {
		t.Fatalf("setup: the attacker is untapped")
	}

	castCatalogSpell(t, g, "Relentless Assault", "Sorcery", relentlessAssaultOracle, nil)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("a creature that attacked this turn is still tapped")
	}
	if !isTapped(g, idle) {
		t.Errorf("a creature that did not attack was untapped")
	}

	next := stepOnce(t, g)
	if next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after Relentless Assault's main phase: %+v, want the second combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attacking again in the added combat: %v", err)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if g.Turn.PhaseOrdinal != 3 {
		t.Errorf("the added main phase reads ordinal %d, want 3 (main phases: precombat, postcombat, added)", g.Turn.PhaseOrdinal)
	}
	if next := stepOnce(t, g); next.Step != game.StepEnd {
		t.Errorf("after the added main phase: %+v, want the end step", next)
	}
}

func TestSeizeTheDayUntapsItsTargetAndAddsCombatAndMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	castCatalogSpell(t, g, "Seize the Day", "Sorcery", seizeTheDayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("Seize the Day's target is still tapped")
	}
	var upcoming []game.PlannedStep
	g.ReadSnapshot(func() { upcoming = g.UpcomingStepsForEffect() })
	if len(upcoming) == 0 || upcoming[0].Step != game.StepBeginCombat || upcoming[0].PhaseID <= 5 {
		t.Fatalf("next step %+v, want the added combat", upcoming)
	}
	// The added main phase follows the added combat, then the turn's
	// own combat.
	var combats, mains int
	for _, p := range upcoming {
		switch p.Step {
		case game.StepBeginCombat:
			combats++
		case game.StepPostcombatMain:
			mains++
		}
	}
	if combats != 2 || mains != 2 {
		t.Errorf("rest of the turn has %d combats and %d postcombat mains, want 2 and 2", combats, mains)
	}
}

func TestAggravatedAssaultActivatesForCombatAndMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	enchantment := pushCatalogPermanent(g, me.ID, "Aggravated Assault", "Enchantment", aggravatedAssaultOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if err := g.ActivateCatalogAbility(me.ID, enchantment, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("a creature you control is still tapped")
	}
	next := stepOnce(t, g)
	if next.Step != game.StepBeginCombat || next.PhaseID <= 5 {
		t.Fatalf("after the activation: %+v, want the added combat", next)
	}
}

func TestFullThrottleAddsTwoCombatsAndUntapsAtEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	castCatalogSpell(t, g, "Full Throttle", "Sorcery", fullThrottleOracle, nil)
	passPriorityAroundTable(t, g)
	const label = "Full Throttle — untap all creatures that attacked this turn"

	var combatIDs []int
	first := stepOnce(t, g)
	combatIDs = append(combatIDs, first.PhaseID)
	if first.Step != game.StepBeginCombat {
		t.Fatalf("after Full Throttle: %+v, want a combat", first)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepEndCombat)
	if !isTapped(g, bear) {
		t.Fatalf("setup: the attacker is untapped after its combat")
	}
	second := stepOnce(t, g)
	if second.Step != game.StepBeginCombat {
		t.Fatalf("after the first added combat: %+v, want the second (no main phase between)", second)
	}
	combatIDs = append(combatIDs, second.PhaseID)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("the beginning of the second combat did not untap the creature that attacked")
	}
	advanceTo(t, g, game.StepEndCombat)
	third := stepOnce(t, g)
	if third.Step != game.StepBeginCombat {
		t.Fatalf("after the second added combat: %+v, want the turn's own combat", third)
	}
	combatIDs = append(combatIDs, third.PhaseID)
	passPriorityAroundTable(t, g)
	if got := countResolved(g, start, label); got != 3 {
		t.Errorf("the each-combat trigger resolved %d times, want 3", got)
	}
	if combatIDs[0] != 6 || combatIDs[1] != 7 || combatIDs[2] != 3 {
		t.Errorf("combats began in phase order %v, want [6 7 3]", combatIDs)
	}
	endTurn(t, g)
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Label == label {
			t.Errorf("the each-combat trigger outlived its turn")
		}
	}
}

func TestKarlachUntapsGrantsFirstStrikeAndAddsOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	karlach := pushDiesCreatureForTest(g, me.ID, "Karlach, Fury of Avernus", karlachOracle, extraPhaseTestLegendaryType, 5, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	const label = "Karlach, Fury of Avernus — untap attackers, first strike, additional combat"

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{karlach, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if countResolved(g, start, label) != 1 {
		t.Fatalf("Karlach's trigger did not resolve once")
	}
	for _, id := range []uuid.UUID{karlach, bear} {
		c, _ := battlefieldCard(g, id)
		if c.Tapped || !game.HasKeyword(&c, "first strike") {
			t.Errorf("%s: tapped=%v, first strike=%v; want untapped with first strike", c.Name, c.Tapped, game.HasKeyword(&c, "first strike"))
		}
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want a second combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(karlach, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, label); n != 1 {
		t.Errorf("Karlach's trigger resolved %d times; the second combat is not the first", n)
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want the postcombat main phase (no third combat)", next)
	}
}

func TestHellkiteChargerPaysForAnAdditionalCombat(t *testing.T) {
	for _, pay := range []bool{true, false} {
		name := "pays"
		if !pay {
			name = "declines"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			hellkite := pushDiesCreatureForTest(g, me.ID, "Hellkite Charger", hellkiteChargerOracle, "Creature — Dragon", 5, 5)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(hellkite, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			lockInAttacks(t, g)
			passPriorityAroundTable(t, g)
			if !hasPayUnlessFor(g, me.ID) {
				t.Fatalf("no pay prompt after Hellkite Charger's trigger")
			}
			// The prompt holds the step: the combat it is about cannot
			// end before it is answered.
			if _, err := g.AdvanceStep(); err == nil && g.Turn.Step != game.StepDeclareAttackers {
				t.Fatalf("the table left the step with the payment unanswered: %+v", g.Turn)
			}
			fillPool(me, 5)
			fillPoolColored(me, "R", 2)
			answerPayUnless(t, g, me.ID, pay)
			if isTapped(g, hellkite) == pay {
				t.Errorf("paid=%v: Hellkite tapped=%v", pay, isTapped(g, hellkite))
			}
			advanceTo(t, g, game.StepEndCombat)
			next := stepOnce(t, g)
			if pay && next.Step != game.StepBeginCombat {
				t.Errorf("paid: after combat %+v, want another combat", next)
			}
			if !pay && next.Step != game.StepPostcombatMain {
				t.Errorf("declined: after combat %+v, want the main phase", next)
			}
		})
	}
}

func TestAureliaUntapsAndAddsACombatOnlyOnHerFirstAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aurelia := pushDiesCreatureForTest(g, me.ID, "Aurelia, the Warleader", aureliaWarleaderOracle, "Legendary Creature — Angel", 3, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	const label = "Aurelia, the Warleader — untap all creatures you control, additional combat"

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{aurelia, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) || isTapped(g, aurelia) {
		t.Errorf("Aurelia's trigger did not untap the creatures you control")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aurelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, label); n != 1 {
		t.Errorf("Aurelia's trigger resolved %d times, want 1 (only her first attack each turn)", n)
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want the main phase", next)
	}
}

func TestSphinxOfTheSecondSunAddsABeginningPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushDiesCreatureForTest(g, me.ID, "Sphinx of the Second Sun", sphinxOfTheSecondSunOracle, "Creature — Sphinx", 6, 6)
	land := pushVanillaCreature(g, me.ID, "Tapped Thing", 1, 1)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	seq := g.Turn.Seq
	hand := me.Hand.Size()

	next := stepOnce(t, g)
	if next.Seq != seq || next.Step != game.StepUpkeep || next.PhaseOrdinal != 2 {
		t.Fatalf("after the postcombat main phase: %+v, want the added beginning phase's upkeep in the same turn", next)
	}
	if isTapped(g, land) {
		t.Errorf("the added untap step did not untap")
	}
	if next := stepOnce(t, g); next.Step != game.StepDraw {
		t.Fatalf("after the added upkeep: %+v, want the draw step", next)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the added draw step draws", me.Hand.Size(), hand+1)
	}
	if next := stepOnce(t, g); next.Step != game.StepEnd || next.Seq != seq {
		t.Errorf("after the added beginning phase: %+v, want this turn's end step", next)
	}
}

func TestYshtolaRhulAddsOneAdditionalEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushDiesCreatureForTest(g, me.ID, "Y'shtola Rhul", yshtolaRhulOracle, "Legendary Creature — Cat Druid", 3, 5)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	const label = "Y'shtola Rhul — blink a creature you control"

	blinkOnce := func(when string) {
		t.Helper()
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		if countResolved(g, start, label) == 0 {
			t.Fatalf("%s: Y'shtola's trigger did not resolve", when)
		}
		// The flicker returns a new object; follow it.
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bear" {
				bear = c.InstanceID
			}
		}
	}
	advanceTo(t, g, game.StepEnd)
	blinkOnce("first end step")
	next := stepOnce(t, g)
	if next.Step != game.StepEnd || next.StepOrdinal != 2 {
		t.Fatalf("after the first end step: %+v, want an additional end step", next)
	}
	blinkOnce("second end step")
	if n := countResolved(g, start, label); n != 2 {
		t.Errorf("Y'shtola's trigger resolved %d times, want 2", n)
	}
	var upcoming []game.PlannedStep
	g.ReadSnapshot(func() { upcoming = g.UpcomingStepsForEffect() })
	if len(upcoming) != 1 || upcoming[0].Step != game.StepCleanup {
		t.Errorf("after the second end step the turn has %+v still to come, want only cleanup", upcoming)
	}
}

func TestEomerAddsACombatWhenAnAttackingLegendDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushDiesCreatureForTest(g, me.ID, "Éomer, Marshal of Rohan", eomerMarshalOfRohanOracle, "Legendary Creature — Human Knight", 4, 4)
	legendA := pushDiesCreatureForTest(g, me.ID, "Legend A", "", extraPhaseTestLegendaryType, 2, 2)
	legendB := pushDiesCreatureForTest(g, me.ID, "Legend B", "", extraPhaseTestLegendaryType, 2, 2)
	grunt := pushVanillaCreature(g, me.ID, "Grunt", 2, 2)
	start := len(g.Events)
	const label = "Éomer, Marshal of Rohan — untap all creatures you control, additional combat"

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{legendA, legendB, grunt} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legendA) })
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, label); n != 1 {
		t.Fatalf("Éomer's trigger resolved %d times, want 1", n)
	}
	if isTapped(g, grunt) {
		t.Errorf("creatures you control were not untapped")
	}
	// Only once each turn.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legendB) })
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, label); n != 1 {
		t.Errorf("Éomer's trigger resolved %d times after a second death, want 1", n)
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after the combat: %+v, want an additional combat", next)
	}
}

// A non-legendary attacker dying does not trigger Éomer.
func TestEomerIgnoresANonLegendaryAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushDiesCreatureForTest(g, me.ID, "Éomer, Marshal of Rohan", eomerMarshalOfRohanOracle, "Legendary Creature — Human Knight", 4, 4)
	grunt := pushVanillaCreature(g, me.ID, "Grunt", 2, 2)
	start := len(g.Events)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(grunt, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(grunt) })
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, "Éomer, Marshal of Rohan — untap all creatures you control, additional combat"); n != 0 {
		t.Errorf("Éomer triggered on a non-legendary creature's death")
	}
}
