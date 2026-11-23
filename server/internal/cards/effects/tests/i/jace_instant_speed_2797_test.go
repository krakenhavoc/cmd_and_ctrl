package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	jaceLine    = "Legendary Planeswalker — Jace"
	chandraLine = "Legendary Planeswalker — Chandra"
)

// Before the spell, a loyalty ability outside the main phase is refused
// (CR 606.3). After it, a Jace's is allowed — in my beginning of combat,
// where only an instant could be cast — and a Chandra's is not.
func TestMachinationsOpensJaceLoyaltyAbilitiesAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, jaceLine, 4)
	chandra := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, chandraLine, 4)

	effectstest.CastCatalogSpell(t, g, "Fixture Machinations", "Instant", effectstest.Pw2797MachinationsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("a Chandra's +2 outside the main phase: got %v, want ErrSorcerySpeedRequired — the statement says Jace", err)
	}
	effectstest.B16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	if got := effectstest.LoyaltyCount(g, jace); got != 6 {
		t.Errorf("the Jace's +2 outside the main phase: loyalty = %d, want 6", got)
	}
	// CR 606.3's other half is untouched: one loyalty ability per permanent
	// per turn, whatever the window.
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
		t.Errorf("a second Jace ability the same turn: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
}

// Without the spell the same activation is refused: the window opens because
// the spell resolved and for no other reason.
func TestJaceLoyaltyAbilityOutsideTheMainPhaseIsRefusedWithoutTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, jaceLine, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("a Jace's +2 in the declare attackers step with no grant: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// "On any player's turn": granted during an opponent's turn (the spell is an
// instant), the statement opens my Jace there, covers only the Jaces I
// control, and ends with that turn.
func TestMachinationsStatementCoversMyJacesOnAnOpponentsTurnAndEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, jaceLine, 4)
	myChandra := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, chandraLine, 4)
	theirs := effectstest.FixtureWalker(g, opp.ID, effectstest.Pw2797WalkerOracle, jaceLine, 4)

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	g.WithWriteLock(func() {
		g.GrantLoyaltyActivationTimingForEffect(me.ID, game.TimingFlash, "Jace", "Fixture Machinations", uuid.Nil, game.Duration{})
	})

	if err := g.ActivateCatalogAbility(me.ID, myChandra, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("my Chandra on the opponent's end step: got %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("their Jace on their own end step: got %v, want ErrSorcerySpeedRequired — the statement is mine", err)
	}
	effectstest.B16Activate(t, g, me.ID, mine, 0, game.ActivateAbilityParams{})
	if got := effectstest.LoyaltyCount(g, mine); got != 6 {
		t.Errorf("my Jace's +2 on the opponent's end step: loyalty = %d, want 6", got)
	}

	// Until end of turn: gone by the next turn's end step, when the permanent
	// has not activated this turn and only the window is in doubt.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("my Jace on a later turn's end step: got %v, want ErrSorcerySpeedRequired — the statement ended with its turn", err)
	}
}

// The statement is plain data on the player, so a restore point keeps it.
func TestMachinationsStatementSurvivesASnapshotRestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := effectstest.FixtureWalker(g, me.ID, effectstest.Pw2797WalkerOracle, jaceLine, 4)
	effectstest.CastCatalogSpell(t, g, "Fixture Machinations", "Instant", effectstest.Pw2797MachinationsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AdvanceTo(t, restored, game.StepDeclareAttackers)
	if err := restored.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the restored game lost the instant-speed statement: %v", err)
	}
}
