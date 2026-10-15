package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// survival_test.go — ADR 0189 slice 1 (#3535): the second-main-phase
// trigger event and the Survival ability word, driven through Defiant
// Survivor, plus the one existing card that needed the same event.

const defiantSurvivorOracle = "7b62c53a-46a2-4c57-9eab-0c6706795b4e"

// survivalTable is a catalog game at the active seat's precombat main
// with a Defiant Survivor on the battlefield and a creature card over a
// land card on top of the library (dreadPair).
func survivalTable(t *testing.T, tapped bool) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, top, second := dreadTable(t)
	id := pushPermanentForTest(g, me.ID, "Defiant Survivor", defiantSurvivorOracle, "Creature — Human Survivor")
	if tapped {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
	}
	return g, me, id, top, second
}

// survivorTriggered reports whether the Survivor has a trigger on the
// stack or waiting to be put there.
func survivorTriggered(g *game.Game, id uuid.UUID) bool {
	if triggerOnStack(g, id) != nil {
		return true
	}
	for _, p := range g.PendingTriggers {
		if p.SourceCardID == id {
			return true
		}
	}
	return false
}

// walkToThirdMain advances to the added third main phase.
func walkToThirdMain(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 40 && !(g.Turn.Step == game.StepPostcombatMain && g.Turn.PhaseOrdinal == 3); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepPostcombatMain || g.Turn.PhaseOrdinal != 3 {
		t.Fatalf("setup: at %s ordinal %d, want the third main phase", g.Turn.Step, g.Turn.PhaseOrdinal)
	}
}

func TestDefiantSurvivorTappedManifestsDreadInTheSecondMainPhase(t *testing.T) {
	g, me, id, top, second := survivalTable(t, true)
	advanceTo(t, g, game.StepPostcombatMain)
	if !survivorTriggered(g, id) {
		t.Fatal("a tapped Defiant Survivor did not trigger at the second main phase")
	}
	answerDread(t, g, me, second)
	passPriorityAroundTable(t, g)
	requireManifested(t, g, me, second)
	if !mdInGraveyard(me, top) {
		t.Error("the card that was not chosen is not in the graveyard")
	}
}

func TestDefiantSurvivorUntappedDoesNotTrigger(t *testing.T) {
	g, me, id, _, _ := survivalTable(t, false)
	advanceTo(t, g, game.StepPostcombatMain)
	if survivorTriggered(g, id) {
		t.Fatal("an untapped Defiant Survivor triggered")
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("manifest dread ran for an untapped Defiant Survivor")
	}
}

func TestDefiantSurvivorUntappedInResponseDoesNothing(t *testing.T) {
	g, me, id, _, _ := survivalTable(t, true)
	advanceTo(t, g, game.StepPostcombatMain)
	if !survivorTriggered(g, id) {
		t.Fatal("setup: no trigger")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("manifest dread ran although the Survivor was untapped on resolution")
	}
}

func TestDefiantSurvivorSacrificedInResponseStillManifestsDread(t *testing.T) {
	g, me, id, _, second := survivalTable(t, true)
	advanceTo(t, g, game.StepPostcombatMain)
	if !survivorTriggered(g, id) {
		t.Fatal("setup: no trigger")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	answerDread(t, g, me, second)
	passPriorityAroundTable(t, g)
	requireManifested(t, g, me, second)
}

func TestDefiantSurvivorTriggersInAMainPhaseAfterASkippedCombat(t *testing.T) {
	g, me, id, _, second := survivalTable(t, true)
	advanceTo(t, g, game.StepBeginCombat)
	g.WithWriteLock(func() { g.EndCombatPhaseForEffect() })
	if g.Turn.Step != game.StepPostcombatMain || g.Turn.PhaseOrdinal != 2 {
		t.Fatalf("setup: at %s ordinal %d, want the second main phase", g.Turn.Step, g.Turn.PhaseOrdinal)
	}
	if !survivorTriggered(g, id) {
		t.Fatal("no trigger in the second main phase of a turn whose combat was cut short")
	}
	answerDread(t, g, me, second)
	passPriorityAroundTable(t, g)
	requireManifested(t, g, me, second)
}

func TestDefiantSurvivorIgnoresAnOpponentsSecondMainPhase(t *testing.T) {
	g, me, id, _, second := survivalTable(t, true)
	mine := g.Turn.ActiveSeat
	// My own second main phase triggers; answer it so the turn can pass.
	advanceTo(t, g, game.StepPostcombatMain)
	answerDread(t, g, me, second)
	passPriorityAroundTable(t, g)
	for i := 0; i < 400 && g.Turn.ActiveSeat == mine; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if g.Turn.ActiveSeat == mine {
		t.Fatal("setup: still my turn")
	}
	if survivorTriggered(g, id) {
		t.Error("Defiant Survivor triggered in an opponent's second main phase")
	}
}

func TestDefiantSurvivorDoesNotTriggerInAnAddedThirdMainPhase(t *testing.T) {
	g, me, id, _, _ := survivalTable(t, false)
	// Untapped through the second main phase, so nothing triggers there;
	// tapped again before the added main phase begins.
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Relentless Assault", "Sorcery", relentlessAssaultOracle, nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
	walkToThirdMain(t, g)
	if !isTapped(g, id) {
		t.Fatal("setup: the Survivor is untapped")
	}
	if survivorTriggered(g, id) {
		t.Error("Defiant Survivor triggered in the added third main phase")
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("manifest dread ran in the third main phase")
	}
}

func TestSchemingSilvertongueDoesNotPrepareInAThirdMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := preparationCard(me.ID, s58SchemingSilvertongue, []string{"flying", "lifelink"},
		game.Face{Name: "Scheming Silvertongue", TypeLine: "Creature — Vampire Warlock", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 1, Toughness: 1},
		game.Face{Name: "Sign in Blood", TypeLine: "Sorcery", ManaCost: "{B}{B}", Colors: []string{"B"}})
	advanceToMain(t, g)
	castFromHandAndResolve(t, g, me, card)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Relentless Assault", "Sorcery", relentlessAssaultOracle, nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(card.InstanceID, me.ID, 2) })
	walkToThirdMain(t, g)
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(card.InstanceID) {
		t.Error("Scheming Silvertongue became prepared in the third main phase")
	}
}
