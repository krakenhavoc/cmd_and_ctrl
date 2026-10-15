package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// savor_gambit_test.go — ADR 0188 Decision 3 (#3563, tracker #753): a
// step skipped in one extra turn (Savor the Moment) and a rules effect
// that lasts during one extra turn (Alchemist's Gambit). Every test
// drives the real cast, the real rotation seam and the real step
// entry, and asserts what the table sees: who untaps, what is
// preventable, who has lost.

const (
	savorTheMomentOracle    = "d5b9b75d-4bf0-4328-9bb3-7b260a0c46af"
	alchemistsGambitOracle  = "271d8db6-8606-46d7-9812-75aa22fcfd8e"
	savorGambitTestLandName = "Tapped Land"
)

// pushTappedLand puts a tapped land under `owner` and returns its ID.
func pushTappedLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: savorGambitTestLandName, TypeLine: "Land",
		Owner: owner, Controller: owner, Tapped: true,
	})
	return id
}

// skipStepRecords counts the skipExtraTurnStep records on the registry.
func skipStepRecords(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModSkipExtraTurnStep {
				n++
			}
		}
	}
	return n
}

// castSavor casts and resolves Savor the Moment for the active seat.
func castSavor(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Savor the Moment", "Sorcery", savorTheMomentOracle, nil)
	passPriorityAroundTable(t, g)
}

// TestSavorTheMomentSkipsTheUntapStepOfTheExtraTurn: the extra turn
// begins at its upkeep, a tapped land stays tapped, and the record is
// spent.
func TestSavorTheMomentSkipsTheUntapStepOfTheExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	land := pushTappedLand(g, g.Seats[me].ID)
	castSavor(t, g)
	if got := skipStepRecords(g); got != 1 {
		t.Fatalf("Savor wrote %d skipExtraTurnStep records, want 1", got)
	}

	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")
	if g.Turn.Step != game.StepUpkeep {
		t.Errorf("the extra turn starts in %q, want the upkeep (the untap step is skipped)", g.Turn.Step)
	}
	if !isTapped(g, land) {
		t.Error("a permanent untapped in the extra turn although its untap step was skipped")
	}
	if got := skipStepRecords(g); got != 0 {
		t.Errorf("the spent skip left %d records behind", got)
	}
}

// TestSavorTheMomentLeavesTheOtherTurnsUntapAlone: the turn the spell
// was cast in already had its untap step, and the turn after the extra
// turn untaps normally. Seat 0 untaps in its next turn.
func TestSavorTheMomentLeavesTheOtherTurnsUntapAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	land := pushTappedLand(g, g.Seats[me].ID)
	castSavor(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")
	if !isTapped(g, land) {
		t.Fatal("setup: the land untapped in the extra turn")
	}
	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
	}
	assertTurnOf(t, g, me, false, "the caster's next normal turn")
	if isTapped(g, land) {
		t.Error("the land did not untap in the caster's next real untap step")
	}
}

// TestSavorTheMomentLeavesAnUntapMarkerForTheNextRealUntapStep:
// "doesn't untap during its controller's next untap step" (CR 614.10a)
// waits for an untap step that happens. The extra turn's is skipped, so
// the marker is spent by the next real one and the land untaps only the
// turn after.
func TestSavorTheMomentLeavesAnUntapMarkerForTheNextRealUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	land := pushTappedLand(g, g.Seats[me].ID)
	g.WithWriteLock(func() {
		if err := g.SkipNextUntapForEffect(land, g.Seats[me].ID); err != nil {
			t.Fatalf("SkipNextUntapForEffect: %v", err)
		}
	})
	castSavor(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")

	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
	}
	assertTurnOf(t, g, me, false, "the first real untap step")
	if !isTapped(g, land) {
		t.Error("the marker was spent by the skipped untap step: the land untapped in the first real one")
	}
	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
	}
	assertTurnOf(t, g, me, false, "the second real untap step")
	if isTapped(g, land) {
		t.Error("the land stayed tapped past the step the marker waited for")
	}
}

// TestSavorTheMomentClearsSummoningSicknessAndSkipsTheDayNightCheck:
// summoning sickness goes as the turn begins, not in the untap step
// (CR 302.6), so last turn's creature can attack; the day/night check is
// part of the untap step (CR 502.2), so it does not run, and the NEXT
// turn's check reads Savor's turn (CR 731.2).
func TestSavorTheMomentClearsSummoningSicknessAndSkipsTheDayNightCheck(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	bear := pushVanillaCreature(g, g.Seats[me].ID, "Fresh Bear", 2, 2)
	g.WithWriteLock(func() {
		g.DayNight.Designation = game.DesignationNight
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
		// The turn Savor was cast in had two spells: a check at the
		// extra turn's untap step would turn the night into day.
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{g.Seats[me].ID: {Total: 2}}
	})
	castSavor(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")

	if !g.IsNight() {
		t.Error("the day/night check ran in the skipped untap step: it is day")
	}
	var sick bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				sick = game.HasSummoningSickness(&g.Battlefield.Cards[i])
			}
		}
	})
	if sick {
		t.Error("the creature is still summoning sick in the extra turn")
	}

	// The extra turn's own spells are what the next check reads.
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{g.Seats[me].ID: {Total: 2}}
	})
	endTurn(t, g)
	if !g.IsDay() {
		t.Error("the next turn's check did not read Savor's turn: still night after two spells in it")
	}
}

// TestSavorTheMomentUnderStasis: Stasis cancels every untap step, and a
// skipped one is spent before its window opens, so the step is skipped
// once, the record is gone, and nothing is asked.
func TestSavorTheMomentUnderStasis(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	_ = seedReplacementPermanent(g, stasisOracle, "Stasis", g.Seats[(me+1)%len(g.Seats)].ID)
	land := pushTappedLand(g, g.Seats[me].ID)
	castSavor(t, g)
	endTurn(t, g)

	assertTurnOf(t, g, me, true, "the extra turn")
	if g.Turn.Step != game.StepUpkeep {
		t.Errorf("the extra turn starts in %q, want the upkeep", g.Turn.Step)
	}
	if !isTapped(g, land) {
		t.Error("the land untapped under Stasis")
	}
	if got := skipStepRecords(g); got != 0 {
		t.Errorf("the skip survived Stasis: %d records", got)
	}
	if n := orderPrompts(g); n != 0 {
		t.Errorf("%d ordering prompts queued, want none", n)
	}
}

// TestSavorTheMomentSkipsOnlyTheFirstUntapStep: Sphinx of the Second Sun
// adds a beginning phase after the extra turn's postcombat main phase,
// and its untap step untaps (owner answer 2 of ADR 0188).
func TestSavorTheMomentSkipsOnlyTheFirstUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castSavor(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")

	pushDiesCreatureForTest(g, g.Seats[me].ID, "Sphinx of the Second Sun", sphinxOfTheSecondSunOracle, "Creature — Sphinx", 6, 6)
	land := pushVanillaCreature(g, g.Seats[me].ID, "Tapped Thing", 1, 1)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })

	next := stepOnce(t, g)
	if next.Step != game.StepUpkeep || next.PhaseOrdinal != 2 {
		t.Fatalf("after the postcombat main phase: %+v, want the added beginning phase's upkeep", next)
	}
	if isTapped(g, land) {
		t.Error("the added beginning phase's untap step was skipped too")
	}
}

// TestSavorTheMomentSkippedByTroubleInPairsLeavesNoRecord: the extra
// turn never begins, so the skip bound to it must not linger (CR 614.10a).
func TestSavorTheMomentSkippedByTroubleInPairsLeavesNoRecord(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	other := (me + 1) % len(g.Seats)
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)
	castSavor(t, g)
	if got := skipStepRecords(g); got != 1 {
		t.Fatalf("Savor wrote %d records, want 1", got)
	}

	endTurn(t, g)
	assertTurnOf(t, g, other, false, "after the skipped extra turn")
	if got := skipStepRecords(g); got != 0 {
		t.Errorf("a skipped extra turn left %d skipExtraTurnStep records", got)
	}
}

// TestAlchemistsGambitPreventionStopsOnlyDuringTheExtraTurn: a Fog cast
// in the turn Gambit resolved in still prevents combat damage; a Fog cast
// in the extra turn prevents nothing (CR 615.12); and the effect ends at
// that turn's cleanup.
func TestAlchemistsGambitPreventionStopsOnlyDuringTheExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Alchemist's Gambit", "Sorcery", alchemistsGambitOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.DamageCantBePreventedThisTurnLabels()) != 0 {
		t.Fatal("the effect began in the turn the spell resolved in")
	}

	// Same turn: Fog still works.
	castCatalogSpell(t, g, "Fog", "Instant", fogOracle, nil)
	passPriorityAroundTable(t, g)
	if got := combatDamageMarked(t, g); got != 0 {
		t.Errorf("Fog in the casting turn left %d damage marked, want 0", got)
	}

	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")
	if len(g.DamageCantBePreventedThisTurnLabels()) != 1 {
		t.Fatalf("the extra turn has %d unpreventable-damage grants, want 1", len(g.DamageCantBePreventedThisTurnLabels()))
	}
	castCatalogSpell(t, g, "Fog", "Instant", fogOracle, nil)
	passPriorityAroundTable(t, g)
	if got := combatDamageMarked(t, g); got != 2 {
		t.Errorf("Fog in the extra turn left %d damage marked, want 2 (it prevents nothing)", got)
	}

	endTurn(t, g)
	if n := len(g.DamageCantBePreventedThisTurnLabels()); n != 0 {
		t.Errorf("the grant outlived the extra turn: %d left", n)
	}
	// ... and the hard-cast loss fired at that turn's end step? It is
	// pinned by TestAlchemistsGambitLosesAtTheExtraTurnsEndStep.
}

// combatDamageMarked deals 2 combat damage from a fresh attacker to a
// fresh 2/2 and returns the damage marked on it.
func combatDamageMarked(t *testing.T, g *game.Game) int {
	t.Helper()
	owner := g.Seats[0].ID
	defenderID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: defenderID, Name: "Defender", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 5, Owner: owner, Controller: owner,
	})
	attackerID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: attackerID, Name: "Attacker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := g.MarkCombatDamage(attackerID, defenderID, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == defenderID {
				marked = c.DamageMarked
			}
		}
	})
	return marked
}

// TestAlchemistsGambitLosesAtTheExtraTurnsEndStep: hard-cast, the
// bracketed loss is a delayed trigger bound to the extra turn.
func TestAlchemistsGambitLosesAtTheExtraTurnsEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	spell := castCatalogSpell(t, g, "Alchemist's Gambit", "Sorcery", alchemistsGambitOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || g.Seats[me].Graveyard.Contains(spell) {
		t.Error("Alchemist's Gambit should exile itself")
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Seats[me].Eliminated {
		t.Fatal("the hard-cast spell's controller did not lose at the extra turn's end step")
	}
}

// TestAlchemistsGambitCleavedDoesNotLose: cleave removes the bracketed
// words, so no loss is scheduled; the damage clause stays.
func TestAlchemistsGambitCleavedDoesNotLose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	active := g.Seats[me]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Alchemist's Gambit", TypeLine: "Sorcery",
		OracleID: alchemistsGambitOracle, Owner: active.ID, Controller: active.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "cleave"}); err != nil {
		t.Fatalf("cleaved Alchemist's Gambit: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("the cleaved cast scheduled %d delayed triggers, want none", len(g.DelayedTriggers))
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn")
	if len(g.DamageCantBePreventedThisTurnLabels()) != 1 {
		t.Errorf("the cleaved cast kept no unpreventable-damage clause")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Seats[me].Eliminated {
		t.Error("the cleaved cast's controller lost the game")
	}
}

// TestAlchemistsGambitSkippedExtraTurnHasNoEffectAndNoLoss: the ruling
// of 2021-11-19 and CR 614.10a. Trouble in Pairs skips the turn: the
// damage clause never starts, the loss is swept, the caster survives.
func TestAlchemistsGambitSkippedExtraTurnHasNoEffectAndNoLoss(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	other := (me + 1) % len(g.Seats)
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)
	castCatalogSpell(t, g, "Alchemist's Gambit", "Sorcery", alchemistsGambitOracle, nil)
	passPriorityAroundTable(t, g)

	endTurn(t, g)
	assertTurnOf(t, g, other, false, "after the skipped extra turn")
	if n := len(g.DamageCantBePreventedThisTurnLabels()); n != 0 {
		t.Errorf("a skipped turn started its effect: %d grants", n)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the loss bound to a skipped turn survived: %d triggers", len(g.DelayedTriggers))
	}
	if g.Seats[me].Eliminated {
		t.Error("the caster lost the game although the extra turn never began")
	}
}
