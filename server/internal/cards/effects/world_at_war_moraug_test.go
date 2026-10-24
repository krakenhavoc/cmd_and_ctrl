package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// world_at_war_moraug_test.go — ADR 0188 slice 6 (#3564): World at War's
// "after the second main phase this turn" and "at the beginning of that
// combat", and Moraug, Fury of Akoum's landfall combat. The engine seams
// are pinned in game/nth_main_phase_test.go; here each card is driven
// through the real cast or land play, and the result is read off where
// the cursor goes and what untaps.

const (
	worldAtWarOracle = "02a8a5cc-62ce-4ae5-aa7d-760124af7d7b"
	moraugOracle     = "6288c0cb-cf32-4ae9-abf0-104ec71654fa"

	worldAtWarUntapLabel = "World at War — untap all creatures that attacked this turn"
	moraugUntapLabel     = "Moraug, Fury of Akoum — untap all creatures you control"
)

// phasesAhead is the phase ids still to come this turn, in order, each
// listed once.
func phasesAhead(g *game.Game) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, p := range g.UpcomingStepsForEffect() {
			if len(out) == 0 || out[len(out)-1] != p.PhaseID {
				out = append(out, p.PhaseID)
			}
		}
	})
	return out
}

// phaseBoundTriggers is the delayed triggers bound to one phase. Rebound
// queues its own upkeep trigger, which is not what these tests count.
func phaseBoundTriggers(g *game.Game) []*game.DelayedTrigger {
	var out []*game.DelayedTrigger
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.OnPhaseID != 0 {
			out = append(out, dt)
		}
	}
	return out
}

func tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// stepIntoCombatOfPhase advances until a beginning of combat in phase id.
func stepIntoCombatOfPhase(t *testing.T, g *game.Game, id int) {
	t.Helper()
	for i := 0; i < 60; i++ {
		if g.Turn.Step == game.StepBeginCombat && g.Turn.PhaseID == id {
			return
		}
		stepOnce(t, g)
	}
	t.Fatalf("never reached the beginning of combat of phase %d (at %+v)", id, g.Turn)
}

func TestWorldAtWarFromThePrecombatMainGoesAfterTheSecondMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	castCatalogSpell(t, g, "World at War", "Sorcery", worldAtWarOracle, nil)
	passPriorityAroundTable(t, g)
	// combat, main, then the pair, then the ending phase.
	if got, want := phasesAhead(g), []int{3, 4, 6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases ahead %v, want %v", got, want)
	}
	bound := phaseBoundTriggers(g)
	if len(bound) != 1 {
		t.Fatalf("%d phase-bound triggers, want the one bound to the added combat", len(bound))
	}
	if dt := bound[0]; dt.OnPhaseID != 6 || dt.At != game.StepBeginCombat {
		t.Errorf("trigger bound to phase %d at %s, want the added combat (6) at its beginning", dt.OnPhaseID, dt.At)
	}
}

func TestTwoWorldAtWarsPutTheNewestPairFirst(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	castCatalogSpell(t, g, "World at War", "Sorcery", worldAtWarOracle, nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "World at War", "Sorcery", worldAtWarOracle, nil)
	passPriorityAroundTable(t, g)
	if got, want := phasesAhead(g), []int{3, 4, 8, 9, 6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases ahead %v, want the second copy's pair (8, 9) before the first's (6, 7)", got)
	}
}

func TestWorldAtWarInTheThirdMainPhaseAddsNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPostcombatMain)
	castCatalogSpell(t, g, "World at War", "Sorcery", worldAtWarOracle, nil)
	passPriorityAroundTable(t, g)
	if got, want := phasesAhead(g), []int{6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases ahead %v, want %v", got, want)
	}
	// Through the added combat (its trigger resolves) to the added main
	// phase, which is the turn's third.
	stepIntoCombatOfPhase(t, g, 6)
	passPriorityAroundTable(t, g)
	for g.Turn.PhaseID != 7 {
		stepOnce(t, g)
	}
	if g.Turn.PhaseOrdinal != 3 {
		t.Fatalf("in main phase %d, want the third", g.Turn.PhaseOrdinal)
	}
	triggers := len(phaseBoundTriggers(g))
	castCatalogSpell(t, g, "World at War", "Sorcery", worldAtWarOracle, nil)
	passPriorityAroundTable(t, g)
	if got, want := phasesAhead(g), []int{5}; !reflect.DeepEqual(got, want) {
		t.Errorf("phases ahead %v after a copy in the third main phase, want only %v", got, want)
	}
	if len(phaseBoundTriggers(g)) != triggers {
		t.Errorf("the copy scheduled an untap for a combat that does not exist")
	}
}

func TestWorldAtWarUntapsAtTheAddedCombatOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	idle := pushVanillaCreature(g, me.ID, "Idle", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	start := len(g.Events)
	castCatalogSpell(t, g, "World at War", "Sorcery", worldAtWarOracle, nil)
	passPriorityAroundTable(t, g)

	// The turn's own combat comes first and does NOT untap.
	stepIntoCombatOfPhase(t, g, 3)
	passPriorityAroundTable(t, g)
	if got := countResolved(g, start, worldAtWarUntapLabel); got != 0 {
		t.Fatalf("the untap resolved %d times at the turn's regular combat, want 0", got)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	tap(g, idle)
	advanceTo(t, g, game.StepEndCombat)
	if !isTapped(g, bear) {
		t.Fatal("setup: the attacker is untapped after the regular combat")
	}

	// The added combat untaps the creature that attacked in the earlier
	// one, and only that one.
	stepIntoCombatOfPhase(t, g, 6)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("a creature that attacked in an earlier combat is still tapped at the added combat")
	}
	if !isTapped(g, idle) {
		t.Errorf("a creature that never attacked was untapped")
	}
	if got := countResolved(g, start, worldAtWarUntapLabel); got != 1 {
		t.Errorf("the untap resolved %d times, want once", got)
	}
}

func TestWorldAtWarReboundedInTheUpkeepAnchorsToThatTurnsSecondMain(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, "World at War", "Sorcery", worldAtWarOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("World at War was not exiled by rebound")
	}

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("no rebound offer at the controller's next upkeep")
	}
	// #4250: the cast below answers the resolution_cast prompt.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the free rebound cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Draw is still ahead in the beginning phase (1), then main (2),
	// combat (3), the second main (4), the pair, the ending phase.
	if got, want := phasesAhead(g), []int{1, 2, 3, 4, 6, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases ahead %v, want %v", got, want)
	}
}

// moraugBoard seeds Moraug and a bear for the active seat.
func moraugBoard(t *testing.T, g *game.Game) (me, opp *game.Player, moraug, bear uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	moraug = pushCatalogPermanent(g, me.ID, "Moraug, Fury of Akoum", "Legendary Creature — Minotaur Warrior", moraugOracle, false)
	bear = pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	return me, opp, moraug, bear
}

func TestMoraugLandfallInThePrecombatMainAddsACombatBeforeTheRegularOne(t *testing.T) {
	g := newCatalogGame(t)
	_, _, _, bear := moraugBoard(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	start := len(g.Events)
	tap(g, bear)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got, want := phasesAhead(g), []int{6, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases ahead %v, want the added combat (6) before the regular one (3): %v", got, want)
	}
	if !isTapped(g, bear) {
		t.Fatal("the untap resolved with the landfall, not at the beginning of the added combat")
	}

	stepIntoCombatOfPhase(t, g, 6)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("the added combat did not untap the creatures")
	}

	// The regular combat does not untap (the 2020-09-25 ruling).
	tap(g, bear)
	stepIntoCombatOfPhase(t, g, 3)
	passPriorityAroundTable(t, g)
	if !isTapped(g, bear) {
		t.Errorf("the turn's regular combat untapped the creatures")
	}
	if got := countResolved(g, start, moraugUntapLabel); got != 1 {
		t.Errorf("the untap resolved %d times, want once", got)
	}
}

func TestMoraugTwoLandfallsInThePostcombatMainMakeTwoConsecutiveCombats(t *testing.T) {
	g := newCatalogGame(t)
	_, _, _, bear := moraugBoard(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	start := len(g.Events)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	// The newest combat is first (CR 500.8), and no main phase sits
	// between them: two consecutive combats, then the ending phase.
	if got, want := phasesAhead(g), []int{7, 6, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("phases ahead %v, want %v", got, want)
	}

	tap(g, bear)
	stepIntoCombatOfPhase(t, g, 7)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("the first added combat did not untap")
	}
	advanceTo(t, g, game.StepEndCombat)
	tap(g, bear)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseID != 6 {
		t.Fatalf("after the first added combat: %+v, want the second added combat at once", next)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Errorf("the second added combat did not untap")
	}
	if got := countResolved(g, start, moraugUntapLabel); got != 2 {
		t.Errorf("the untap resolved %d times, want once per added combat", got)
	}
}

// The anthem reads 0, 1, then 2 declarations across two combats, and a
// creature put onto the battlefield attacking is not counted (the
// 2020-09-25 ruling: it hasn't attacked).
func TestMoraugPumpCountsDeclarationsAcrossTwoCombats(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _, bear := moraugBoard(t, g)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Fatalf("before attacking power %d, want 2", got)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if n, p := timesAttacked(g, bear), effectivePower(t, g, bear); n != 1 || p != 3 {
		t.Fatalf("after one attack: %d declarations, power %d, want 1 and 3", n, p)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	stepIntoCombatOfPhase(t, g, 6)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Fatal("setup: the added combat did not untap the bear")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers in the added combat: %v", err)
	}
	if n, p := timesAttacked(g, bear), effectivePower(t, g, bear); n != 2 || p != 4 {
		t.Fatalf("after two attacks: %d declarations, power %d, want 2 and 4", n, p)
	}

	// A creature that enters attacking was never declared.
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		err := g.CreateTokensAttackingForEffect(me.ID,
			game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}, 1, opp.ID)
		if err != nil {
			t.Fatalf("CreateTokensAttackingForEffect: %v", err)
		}
	})
	tok := battlefieldIDsNamed(g, "Goblin")
	if len(tok) != 1 {
		t.Fatalf("%d Goblin tokens, want 1", len(tok))
	}
	if n, p := timesAttacked(g, tok[0]), effectivePower(t, g, tok[0]); n != 0 || p != 1 {
		t.Errorf("a creature put onto the battlefield attacking: %d declarations, power %d, want 0 and 1", n, p)
	}
}

func timesAttacked(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() { n = g.TimesAttackedThisTurn(id) })
	return n
}

// CR 903.10a: commander damage is cumulative for the game, so a commander
// that attacks in both combats under Moraug adds up. 6 power plus 1 for
// the first declaration is 7; the second combat reads 8; the tally is 15.
func TestMoraugCommanderDamageAddsUpAcrossBothCombats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _, _ := moraugBoard(t, g)
	cmdr := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cmdr, Name: "Big Commander", TypeLine: "Legendary Creature — Test",
		Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cmdr, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := opp.CommanderDamage[cmdr]; got != 7 {
		t.Fatalf("commander damage after the first combat %d, want 7", got)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	stepIntoCombatOfPhase(t, g, 6)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cmdr, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the added combat: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := opp.CommanderDamage[cmdr]; got != 15 {
		t.Errorf("commander damage after both combats %d, want 15 (7 + 8)", got)
	}
}
