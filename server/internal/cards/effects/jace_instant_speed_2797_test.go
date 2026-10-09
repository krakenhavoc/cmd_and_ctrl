package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jace_instant_speed_2797_test.go — #2797: "Until end of turn, you may
// activate loyalty abilities of Jace planeswalkers you control on any
// player's turn any time you could cast an instant" (Jace's Machinations),
// proved on a fixture instant and fixture planeswalkers so it does not wait
// on "empower Jace" (#2796).
//
// The Teferi helpers state this for a source (Teferi, Master of Time) or an
// emblem (Teferi, Temporal Archmage); neither lasts "until end of turn". The
// resolving-spell form is a stored statement on the player with a duration
// (game/activation_timing_grant.go).

const pw2797MachinationsOracle = "test-2797-machinations"

func init() {
	Register(Spec{
		OracleID: pw2797MachinationsOracle,
		Name:     "Fixture Machinations",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantLoyaltyAbilitiesAtInstantSpeed{
				Subtype: "Jace",
				Label:   "Fixture Machinations — loyalty abilities of Jace planeswalkers at instant speed",
			}.Apply(ctx)
		},
	})
}

const (
	jaceLine    = "Legendary Planeswalker — Jace"
	chandraLine = "Legendary Planeswalker — Chandra"
)

// Before the spell, a loyalty ability outside the main phase is refused
// (CR 606.3). After it, a Jace's is allowed — in my beginning of combat,
// where only an instant could be cast — and a Chandra's is not.
func TestMachinationsOpensJaceLoyaltyAbilitiesAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := fixtureWalker(g, me.ID, pw2797WalkerOracle, jaceLine, 4)
	chandra := fixtureWalker(g, me.ID, pw2797WalkerOracle, chandraLine, 4)

	castCatalogSpell(t, g, "Fixture Machinations", "Instant", pw2797MachinationsOracle, nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)

	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("a Chandra's +2 outside the main phase: got %v, want ErrSorcerySpeedRequired — the statement says Jace", err)
	}
	b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	if got := loyaltyCount(g, jace); got != 6 {
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
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := fixtureWalker(g, me.ID, pw2797WalkerOracle, jaceLine, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("a Jace's +2 in the declare attackers step with no grant: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// "On any player's turn": granted during an opponent's turn (the spell is an
// instant), the statement opens my Jace there, covers only the Jaces I
// control, and ends with that turn.
func TestMachinationsStatementCoversMyJacesOnAnOpponentsTurnAndEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := fixtureWalker(g, me.ID, pw2797WalkerOracle, jaceLine, 4)
	myChandra := fixtureWalker(g, me.ID, pw2797WalkerOracle, chandraLine, 4)
	theirs := fixtureWalker(g, opp.ID, pw2797WalkerOracle, jaceLine, 4)

	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepEnd)
	g.WithWriteLock(func() {
		g.GrantLoyaltyActivationTimingForEffect(me.ID, game.TimingFlash, "Jace", "Fixture Machinations", uuid.Nil, game.Duration{})
	})

	if err := g.ActivateCatalogAbility(me.ID, myChandra, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("my Chandra on the opponent's end step: got %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("their Jace on their own end step: got %v, want ErrSorcerySpeedRequired — the statement is mine", err)
	}
	b16Activate(t, g, me.ID, mine, 0, game.ActivateAbilityParams{})
	if got := loyaltyCount(g, mine); got != 6 {
		t.Errorf("my Jace's +2 on the opponent's end step: loyalty = %d, want 6", got)
	}

	// Until end of turn: gone by the next turn's end step, when the permanent
	// has not activated this turn and only the window is in doubt.
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepEnd)
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("my Jace on a later turn's end step: got %v, want ErrSorcerySpeedRequired — the statement ended with its turn", err)
	}
}

// The statement is plain data on the player, so a restore point keeps it.
func TestMachinationsStatementSurvivesASnapshotRestore(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := fixtureWalker(g, me.ID, pw2797WalkerOracle, jaceLine, 4)
	castCatalogSpell(t, g, "Fixture Machinations", "Instant", pw2797MachinationsOracle, nil)
	passPriorityAroundTable(t, g)

	restored := restoreRoundTrip(t, g, true)
	advanceTo(t, restored, game.StepDeclareAttackers)
	if err := restored.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the restored game lost the instant-speed statement: %v", err)
	}
}
