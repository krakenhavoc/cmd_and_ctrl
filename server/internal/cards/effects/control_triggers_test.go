package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// control_triggers_test.go — the three printed shapes #930 adds, each
// on its own probe card so one event cannot satisfy two of them at
// once. The engine half (one event per delta, from the one
// materialise step) is pinned in game/control_changed_event_test.go.

const (
	loseControlProbeOracle  = "test-control-lose-probe"
	gainControlProbeOracle  = "test-control-gain-probe"
	ownedControlProbeOracle = "test-control-owned-probe"
)

func init() {
	Register(Spec{
		OracleID: loseControlProbeOracle,
		Name:     "Betrayer Probe",
		Triggered: []game.TriggeredAbility{
			WhenYouLoseControlOfThis("Betrayer Probe — draw two cards", Do(DrawCards{N: 2})),
		},
	})
	Register(Spec{
		OracleID: gainControlProbeOracle,
		Name:     "Turncoat Probe",
		Triggered: []game.TriggeredAbility{
			WhenYouGainControlOfThis("Turncoat Probe — gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
	Register(Spec{
		OracleID: ownedControlProbeOracle,
		Name:     "Donor Probe",
		Triggered: []game.TriggeredAbility{
			WheneverAnOpponentGainsControlOfAPermanentYouOwn(
				"Donor Probe — gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}

// stealForTest hands `target` to `to` until end of turn and settles
// the layer pass, which is where the control change — and its event —
// happens.
func stealForTest(t *testing.T, g *game.Game, target, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), target, to, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatal("GainControlForEffect refused a battlefield permanent")
		}
	})
	g.ReadSnapshot(func() {})
}

// pendingTriggerFor returns the harvested trigger queued for
// `cardID`, before the next priority boundary drains it onto the
// stack. The queue is where a control-change trigger lands: the
// event is emitted from the layer pass, not from a resolution.
func pendingTriggerFor(g *game.Game, cardID uuid.UUID) *game.StackItem {
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == cardID {
			return item
		}
	}
	return nil
}

// TestWhenYouLoseControlOfThisFiresForThePlayerWhoLostIt — Khârn the
// Betrayer's Sigil of Corruption. "You" is the player the permanent
// was taken FROM, so the item on the stack is theirs even though the
// permanent is not any more.
func TestWhenYouLoseControlOfThisFiresForThePlayerWhoLostIt(t *testing.T) {
	g := newCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	probe := pushCatalogPermanent(g, me.ID, "Betrayer Probe", "Creature — Test", loseControlProbeOracle, false)
	g.ReadSnapshot(func() {})
	before := me.Hand.Size()
	thiefBefore := thief.Hand.Size()

	stealForTest(t, g, probe, thief.ID)

	item := pendingTriggerFor(g, probe)
	if item == nil {
		t.Fatal("no trigger queued after the permanent changed control")
	}
	if item.Controller != me.ID {
		t.Errorf("trigger controller %s, want the player who LOST control %s", item.Controller, me.ID)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("the player who lost control has %d cards, want %d", got, before+2)
	}
	if got := thief.Hand.Size(); got != thiefBefore {
		t.Errorf("the thief drew %d cards; the trigger is not theirs", got-thiefBefore)
	}
}

// TestWhenYouLoseControlOfThisFiresAgainWhenTheTheftEnds — the revert
// at the duration's expiry is a control change like any other, so the
// thief loses control and the trigger fires for THEM. Once per change,
// not once per turn.
func TestWhenYouLoseControlOfThisFiresAgainWhenTheTheftEnds(t *testing.T) {
	g := newCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	probe := pushCatalogPermanent(g, me.ID, "Betrayer Probe", "Creature — Test", loseControlProbeOracle, false)
	g.ReadSnapshot(func() {})

	stealForTest(t, g, probe, thief.ID)
	passPriorityAroundTable(t, g)
	thiefBefore := thief.Hand.Size()

	advancePastCleanupForTest(t, g)
	passPriorityAroundTable(t, g)

	if got := thief.Hand.Size(); got != thiefBefore+2 {
		t.Errorf("the thief has %d cards after the theft expired, want %d — losing control at a "+
			"duration's end is still losing control", got, thiefBefore+2)
	}
}

// TestWhenYouGainControlOfThisFiresForTheNewController — Risky Move's
// half. The gaining player already controls the permanent by the time
// the event lands, so the ordinary item is theirs.
func TestWhenYouGainControlOfThisFiresForTheNewController(t *testing.T) {
	g := newCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	probe := pushCatalogPermanent(g, me.ID, "Turncoat Probe", "Creature — Test", gainControlProbeOracle, false)
	g.ReadSnapshot(func() {})
	before := thief.Life

	stealForTest(t, g, probe, thief.ID)

	item := pendingTriggerFor(g, probe)
	if item == nil {
		t.Fatal("no trigger queued after the permanent changed control")
	}
	if item.Controller != thief.ID {
		t.Errorf("trigger controller %s, want the player who GAINED control %s", item.Controller, thief.ID)
	}
	passPriorityAroundTable(t, g)

	if thief.Life != before+3 {
		t.Errorf("the new controller is at %d life, want %d", thief.Life, before+3)
	}
}

// TestWheneverAnOpponentGainsControlOfAPermanentYouOwn — the Zedruu
// shape: a watcher permanent, a different permanent that moved, and
// OWNERSHIP as the link (CR 108.3).
func TestWheneverAnOpponentGainsControlOfAPermanentYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Donor Probe", "Creature — Test", ownedControlProbeOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, thief.ID, "Bear", "Creature — Bear", "", false)
	g.ReadSnapshot(func() {})
	before := me.Life

	// A permanent I own, taken by an opponent: the trigger.
	stealForTest(t, g, mine, thief.ID)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("life %d after an opponent took a permanent I own, want %d", me.Life, before+1)
	}

	// A permanent the opponent owns, taken by ME: not my trigger, and
	// not an opponent gaining anything.
	stealForTest(t, g, theirs, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after I took a permanent an opponent owns, want %d — the clause reads "+
			"ownership and names an opponent as the gaining player", me.Life, before+1)
	}
}
