package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRelentlessAssaultUntapsAttackersAndAddsCombatAndMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(idle) })
	if !effectstest.IsTapped(g, bear) {
		t.Fatalf("setup: the attacker is untapped")
	}

	effectstest.CastCatalogSpell(t, g, "Relentless Assault", "Sorcery", effectstest.RelentlessAssaultOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Errorf("a creature that attacked this turn is still tapped")
	}
	if !effectstest.IsTapped(g, idle) {
		t.Errorf("a creature that did not attack was untapped")
	}

	next := effectstest.StepOnce(t, g)
	if next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after Relentless Assault's main phase: %+v, want the second combat", next)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attacking again in the added combat: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if g.Turn.PhaseOrdinal != 3 {
		t.Errorf("the added main phase reads ordinal %d, want 3 (main phases: precombat, postcombat, added)", g.Turn.PhaseOrdinal)
	}
	if next := effectstest.StepOnce(t, g); next.Step != game.StepEnd {
		t.Errorf("after the added main phase: %+v, want the end step", next)
	}
}

func TestSeizeTheDayUntapsItsTargetAndAddsCombatAndMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	effectstest.CastCatalogSpell(t, g, "Seize the Day", "Sorcery", effectstest.SeizeTheDayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	enchantment := effectstest.PushCatalogPermanent(g, me.ID, "Aggravated Assault", "Enchantment", effectstest.AggravatedAssaultOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if err := g.ActivateCatalogAbility(me.ID, enchantment, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Errorf("a creature you control is still tapped")
	}
	next := effectstest.StepOnce(t, g)
	if next.Step != game.StepBeginCombat || next.PhaseID <= 5 {
		t.Fatalf("after the activation: %+v, want the added combat", next)
	}
}

func TestFullThrottleAddsTwoCombatsAndUntapsAtEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	effectstest.CastCatalogSpell(t, g, "Full Throttle", "Sorcery", effectstest.FullThrottleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	const label = "Full Throttle — untap all creatures that attacked this turn"

	var combatIDs []int
	first := effectstest.StepOnce(t, g)
	combatIDs = append(combatIDs, first.PhaseID)
	if first.Step != game.StepBeginCombat {
		t.Fatalf("after Full Throttle: %+v, want a combat", first)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if !effectstest.IsTapped(g, bear) {
		t.Fatalf("setup: the attacker is untapped after its combat")
	}
	second := effectstest.StepOnce(t, g)
	if second.Step != game.StepBeginCombat {
		t.Fatalf("after the first added combat: %+v, want the second (no main phase between)", second)
	}
	combatIDs = append(combatIDs, second.PhaseID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Errorf("the beginning of the second combat did not untap the creature that attacked")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	third := effectstest.StepOnce(t, g)
	if third.Step != game.StepBeginCombat {
		t.Fatalf("after the second added combat: %+v, want the turn's own combat", third)
	}
	combatIDs = append(combatIDs, third.PhaseID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountResolved(g, start, label); got != 3 {
		t.Errorf("the each-combat trigger resolved %d times, want 3", got)
	}
	if combatIDs[0] != 6 || combatIDs[1] != 7 || combatIDs[2] != 3 {
		t.Errorf("combats began in phase order %v, want [6 7 3]", combatIDs)
	}
	effectstest.EndTurn(t, g)
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Label == label {
			t.Errorf("the each-combat trigger outlived its turn")
		}
	}
}

func TestKarlachUntapsGrantsFirstStrikeAndAddsOneCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	karlach := effectstest.PushDiesCreatureForTest(g, me.ID, "Karlach, Fury of Avernus", effectstest.KarlachOracle, effectstest.ExtraPhaseTestLegendaryType, 5, 4)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	const label = "Karlach, Fury of Avernus — untap attackers, first strike, additional combat"

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{karlach, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountResolved(g, start, label) != 1 {
		t.Fatalf("Karlach's trigger did not resolve once")
	}
	for _, id := range []uuid.UUID{karlach, bear} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Tapped || !game.HasKeyword(&c, "first strike") {
			t.Errorf("%s: tapped=%v, first strike=%v; want untapped with first strike", c.Name, c.Tapped, game.HasKeyword(&c, "first strike"))
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want a second combat", next)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(karlach, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, label); n != 1 {
		t.Errorf("Karlach's trigger resolved %d times; the second combat is not the first", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepPostcombatMain {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			hellkite := effectstest.PushDiesCreatureForTest(g, me.ID, "Hellkite Charger", effectstest.HellkiteChargerOracle, "Creature — Dragon", 5, 5)
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(hellkite, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			effectstest.LockInAttacks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatalf("no pay prompt after Hellkite Charger's trigger")
			}
			// The prompt holds the step: the combat it is about cannot
			// end before it is answered.
			if _, err := g.AdvanceStep(); err == nil && g.Turn.Step != game.StepDeclareAttackers {
				t.Fatalf("the table left the step with the payment unanswered: %+v", g.Turn)
			}
			effectstest.FillPool(me, 5)
			effectstest.FillPoolColored(me, "R", 2)
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
			if effectstest.IsTapped(g, hellkite) == pay {
				t.Errorf("paid=%v: Hellkite tapped=%v", pay, effectstest.IsTapped(g, hellkite))
			}
			effectstest.AdvanceTo(t, g, game.StepEndCombat)
			next := effectstest.StepOnce(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aurelia := effectstest.PushDiesCreatureForTest(g, me.ID, "Aurelia, the Warleader", effectstest.AureliaWarleaderOracle, "Legendary Creature — Angel", 3, 4)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	const label = "Aurelia, the Warleader — untap all creatures you control, additional combat"

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{aurelia, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) || effectstest.IsTapped(g, aurelia) {
		t.Errorf("Aurelia's trigger did not untap the creatures you control")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aurelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, label); n != 1 {
		t.Errorf("Aurelia's trigger resolved %d times, want 1 (only her first attack each turn)", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want the main phase", next)
	}
}

func TestSphinxOfTheSecondSunAddsABeginningPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Sphinx of the Second Sun", effectstest.SphinxOfTheSecondSunOracle, "Creature — Sphinx", 6, 6)
	land := effectstest.PushVanillaCreature(g, me.ID, "Tapped Thing", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	seq := g.Turn.Seq
	hand := me.Hand.Size()

	next := effectstest.StepOnce(t, g)
	if next.Seq != seq || next.Step != game.StepUpkeep || next.PhaseOrdinal != 2 {
		t.Fatalf("after the postcombat main phase: %+v, want the added beginning phase's upkeep in the same turn", next)
	}
	if effectstest.IsTapped(g, land) {
		t.Errorf("the added untap step did not untap")
	}
	if next := effectstest.StepOnce(t, g); next.Step != game.StepDraw {
		t.Fatalf("after the added upkeep: %+v, want the draw step", next)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the added draw step draws", me.Hand.Size(), hand+1)
	}
	if next := effectstest.StepOnce(t, g); next.Step != game.StepEnd || next.Seq != seq {
		t.Errorf("after the added beginning phase: %+v, want this turn's end step", next)
	}
}

func TestYshtolaRhulAddsOneAdditionalEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Y'shtola Rhul", effectstest.YshtolaRhulOracle, "Legendary Creature — Cat Druid", 3, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)
	const label = "Y'shtola Rhul — blink a creature you control"

	blinkOnce := func(when string) {
		t.Helper()
		effectstest.AnswerPickTarget(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.CountResolved(g, start, label) == 0 {
			t.Fatalf("%s: Y'shtola's trigger did not resolve", when)
		}
		// The flicker returns a new object; follow it.
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bear" {
				bear = c.InstanceID
			}
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	blinkOnce("first end step")
	next := effectstest.StepOnce(t, g)
	if next.Step != game.StepEnd || next.StepOrdinal != 2 {
		t.Fatalf("after the first end step: %+v, want an additional end step", next)
	}
	blinkOnce("second end step")
	if n := effectstest.CountResolved(g, start, label); n != 2 {
		t.Errorf("Y'shtola's trigger resolved %d times, want 2", n)
	}
	var upcoming []game.PlannedStep
	g.ReadSnapshot(func() { upcoming = g.UpcomingStepsForEffect() })
	if len(upcoming) != 1 || upcoming[0].Step != game.StepCleanup {
		t.Errorf("after the second end step the turn has %+v still to come, want only cleanup", upcoming)
	}
}

func TestEomerAddsACombatWhenAnAttackingLegendDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Éomer, Marshal of Rohan", effectstest.EomerMarshalOfRohanOracle, "Legendary Creature — Human Knight", 4, 4)
	legendA := effectstest.PushDiesCreatureForTest(g, me.ID, "Legend A", "", effectstest.ExtraPhaseTestLegendaryType, 2, 2)
	legendB := effectstest.PushDiesCreatureForTest(g, me.ID, "Legend B", "", effectstest.ExtraPhaseTestLegendaryType, 2, 2)
	grunt := effectstest.PushVanillaCreature(g, me.ID, "Grunt", 2, 2)
	start := len(g.Events)
	const label = "Éomer, Marshal of Rohan — untap all creatures you control, additional combat"

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{legendA, legendB, grunt} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legendA) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, label); n != 1 {
		t.Fatalf("Éomer's trigger resolved %d times, want 1", n)
	}
	if effectstest.IsTapped(g, grunt) {
		t.Errorf("creatures you control were not untapped")
	}
	// Only once each turn.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legendB) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, label); n != 1 {
		t.Errorf("Éomer's trigger resolved %d times after a second death, want 1", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after the combat: %+v, want an additional combat", next)
	}
}

// A non-legendary attacker dying does not trigger Éomer.
func TestEomerIgnoresANonLegendaryAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Éomer, Marshal of Rohan", effectstest.EomerMarshalOfRohanOracle, "Legendary Creature — Human Knight", 4, 4)
	grunt := effectstest.PushVanillaCreature(g, me.ID, "Grunt", 2, 2)
	start := len(g.Events)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(grunt, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(grunt) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, "Éomer, Marshal of Rohan — untap all creatures you control, additional combat"); n != 0 {
		t.Errorf("Éomer triggered on a non-legendary creature's death")
	}
}
