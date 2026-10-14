package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_trigger_controller_test.go — DelayedTrigger.ManaTapControllerOnly
// (#3504): "until end of turn, whenever YOU tap a Mountain for mana, add
// an additional {R}" (Pyre Rhymer // Molten Tide). The unfiltered twin,
// "whenever a player taps" (High Tide, Bubbling Muck), is pinned by
// TestTurnManaTriggerAddsForAnyPlayersSwampThisTurn and must not change.

func scheduleTurnMana(g *Game, controller uuid.UUID, controllerOnly bool) {
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:            controller,
			SourceCardID:          uuid.New(),
			Label:                 "test — add an additional {R}",
			ManaTapSubtype:        "Mountain",
			ManaAdds:              "{R}",
			Repeats:               true,
			ManaTapControllerOnly: controllerOnly,
		})
	})
}

func TestControllerOnlyManaTriggerFiresForYourMountainsOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scheduleTurnMana(g, me.ID, true)
	mine := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "test-mountain")
	theirs := pushBattlefieldForTest(g, opp.ID, "Mountain", "Basic Land — Mountain", "test-mountain")
	forest := pushForest(g, me)

	if err := g.ActivateManaAbility(me.ID, mine, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); got["R"] != 2 {
		t.Errorf("my pool = %v, want the Mountain's {R} and the extra", got)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(opp); got["R"] != 1 {
		t.Errorf("their pool = %v, want only their Mountain's own {R}: \"you\" is the caster", got)
	}
	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); got["G"] != 1 || got["R"] != 2 {
		t.Errorf("my pool after a Forest = %v, want no extra for a non-Mountain", got)
	}
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Errorf("a mana trigger used the stack: %d pending, %d meta", len(g.PendingTriggers), len(g.StackMeta))
	}
}

// The same record without the flag still fires for every player: the
// default is the "a player" reading High Tide and Bubbling Muck print.
func TestUnfilteredManaTriggerStillFiresForEveryPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scheduleTurnMana(g, me.ID, false)
	theirs := pushBattlefieldForTest(g, opp.ID, "Mountain", "Basic Land — Mountain", "test-mountain")

	if err := g.ActivateManaAbility(opp.ID, theirs, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(opp); got["R"] != 2 {
		t.Errorf("their pool = %v, want the extra {R} for any player's Mountain", got)
	}
}

func TestControllerOnlyManaTriggerSurvivesSnapshotAndClone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scheduleTurnMana(g, me.ID, true)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game owing a controller-only mana trigger is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.DelayedTriggers) != 1 || !restored.DelayedTriggers[0].ManaTapControllerOnly {
		t.Fatalf("restored queue = %+v, want one record still controller-only", restored.DelayedTriggers)
	}
	cl := g.Clone()
	if len(cl.DelayedTriggers) != 1 || !cl.DelayedTriggers[0].ManaTapControllerOnly {
		t.Fatalf("Clone lost ManaTapControllerOnly: %+v", cl.DelayedTriggers)
	}

	theirs := pushBattlefieldForTest(restored, opp.ID, "Mountain", "Basic Land — Mountain", "test-mountain")
	if err := restored.ActivateManaAbility(opp.ID, theirs, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(restored.Seats[1]); got["R"] != 1 {
		t.Errorf("restored: their pool = %v, want no extra for the opponent", got)
	}
}
