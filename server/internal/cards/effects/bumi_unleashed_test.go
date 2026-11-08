package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bumi_unleashed_test.go — Bumi, Unleashed (#4459) and the seam it waited
// on (#5277, ADR 0188 Decision 13): "only land creatures can attack
// during that combat phase" is in force in the combat Bumi adds and in no
// other.

const bumiUnleashedOracle = "1fef13d8-5da2-46a6-91ed-7a31bd46e538"

func TestBumiUnleashedIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(bumiUnleashedOracle)
	if !ok || spec.Name != "Bumi, Unleashed" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Bumi, Unleashed: ok %v, %+v", ok, spec)
	}
}

// bumiScene is Bumi, two creatures that aren't lands, a land creature
// and a tapped land, all under the active player.
type bumiScene struct {
	g                   *game.Game
	me, opp             *game.Player
	bumi, bear1, bear2  uuid.UUID
	arbor, tappedForest uuid.UUID
}

func newBumiScene(t *testing.T) bumiScene {
	t.Helper()
	g := newCatalogGame(t)
	s := bumiScene{g: g, me: g.Seats[0], opp: g.Seats[1]}
	s.bumi = b12Push(g, s.me.ID, "Bumi, Unleashed", "Legendary Creature — Human Noble Ally", bumiUnleashedOracle, 5, 4)
	s.bear1 = b12Creature(g, s.me.ID, "Bear One", "Creature — Bear", 2, 2)
	s.bear2 = b12Creature(g, s.me.ID, "Bear Two", "Creature — Bear", 2, 2)
	s.arbor = b12Creature(g, s.me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", 1, 1)
	s.tappedForest = b12Permanent(g, s.me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(s.tappedForest) })
	return s
}

// connect attacks with Bumi, Bear One and the land creature in the
// turn's first combat (nothing restricts that combat) and resolves
// Bumi's combat-damage trigger.
func (s bumiScene) connect(t *testing.T) {
	t.Helper()
	attackWith(t, s.g, s.opp.ID, s.bumi, s.bear1, s.arbor)
	passPriorityAroundTable(t, s.g)
	if len(s.g.PhaseEffects) != 1 {
		t.Fatalf("Bumi's trigger left %d waiting rules, want 1", len(s.g.PhaseEffects))
	}
}

func cantAttackNow(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return restrictionsOf(t, g, id)&game.CantAttack != 0
}

func TestBumiUnleashedUntapsLandsAndOnlyLandCreaturesAttackInItsCombat(t *testing.T) {
	s := newBumiScene(t)
	first := s.g.Turn.PhaseID
	s.connect(t)
	if tappedNow(s.g, s.tappedForest) || tappedNow(s.g, s.arbor) {
		t.Error("Bumi did not untap every land you control, the land creature included")
	}
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if s.g.Turn.PhaseID == first {
		t.Fatal("no combat was added")
	}
	if err := s.g.DeclareAttacker(s.bear2, s.opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("a creature that isn't a land attacked in Bumi's combat: err = %v", err)
	}
	if err := s.g.DeclareAttacker(s.arbor, s.opp.ID); err != nil {
		t.Errorf("the land creature could not attack in Bumi's combat: %v", err)
	}
}

func TestBumiUnleashedFirstCombatIsUnrestricted(t *testing.T) {
	s := newBumiScene(t)
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if cantAttackNow(t, s.g, s.bear1) {
		t.Fatal("a creature is restricted in the first combat")
	}
	// Bear One attacks in the first combat beside Bumi.
	s.connect(t)
	if !s.g.AttackedThisTurn(s.bear1) {
		t.Error("Bear One did not attack in the first combat")
	}
}

// The 2025-10-02 ruling, with CR 500.8: a combat added after Bumi's
// trigger resolved, in the same phase, runs first and is not restricted.
// Bumi's own combat follows it and is.
func TestBumiUnleashedANewerAddedCombatThatRunsFirstIsUnrestricted(t *testing.T) {
	s := newBumiScene(t)
	s.connect(t)
	s.g.WithWriteLock(func() {
		s.g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})
	advanceTo(t, s.g, game.StepDeclareAttackers)
	newer := s.g.Turn.PhaseID
	if cantAttackNow(t, s.g, s.bear2) {
		t.Fatal("the newer added combat, which runs first, is restricted")
	}
	advanceTo(t, s.g, game.StepEndCombat)
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if s.g.Turn.PhaseID == newer {
		t.Fatal("never reached Bumi's combat")
	}
	if !cantAttackNow(t, s.g, s.bear2) {
		t.Error("Bumi's combat is not restricted after a newer combat ran first")
	}
	if cantAttackNow(t, s.g, s.arbor) {
		t.Error("the land creature is restricted in Bumi's combat")
	}
}

// A restore point taken while the rule waits, and one taken in Bumi's
// combat, both keep it.
func TestBumiUnleashedRuleSurvivesARestoreMidCombat(t *testing.T) {
	s := newBumiScene(t)
	s.connect(t)
	waiting, err := s.g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore while the rule waits: %v", err)
	}
	if len(waiting.PhaseEffects) != 1 {
		t.Fatalf("restored %d waiting rules, want 1", len(waiting.PhaseEffects))
	}
	advanceTo(t, waiting, game.StepDeclareAttackers)
	if !cantAttackNow(t, waiting, s.bear2) {
		t.Error("the restored game's added combat is not restricted")
	}

	advanceTo(t, s.g, game.StepDeclareAttackers)
	inCombat, err := s.g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore in Bumi's combat: %v", err)
	}
	if err := inCombat.DeclareAttacker(s.bear2, s.opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("after a restore in Bumi's combat a creature that isn't a land attacked: err = %v", err)
	}
}

func TestBumiUnleashedEarthbendsFourOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Bumi, Unleashed", "Legendary Creature — Human Noble Ally", bumiUnleashedOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, land, game.CounterPlusOne); got != 4 {
		t.Errorf("the land has %d +1/+1 counters, want 4", got)
	}
	if v := s58p6View(t, g, land); !v.IsCreature() || !v.IsLand() {
		t.Errorf("the earthbent land: creature %v, land %v; want a land creature", v.IsCreature(), v.IsLand())
	}
}
