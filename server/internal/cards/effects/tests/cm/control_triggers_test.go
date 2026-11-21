package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// TestWhenYouLoseControlOfThisFiresForThePlayerWhoLostIt — Khârn the
// Betrayer's Sigil of Corruption. "You" is the player the permanent
// was taken FROM, so the item on the stack is theirs even though the
// permanent is not any more.
func TestWhenYouLoseControlOfThisFiresForThePlayerWhoLostIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	probe := effectstest.PushCatalogPermanent(g, me.ID, "Betrayer Probe", "Creature — Test", effectstest.LoseControlProbeOracle, false)
	g.ReadSnapshot(func() {})
	before := me.Hand.Size()
	thiefBefore := thief.Hand.Size()

	effectstest.StealForTest(t, g, probe, thief.ID)

	item := effectstest.PendingTriggerFor(g, probe)
	if item == nil {
		t.Fatal("no trigger queued after the permanent changed control")
	}
	if item.Controller != me.ID {
		t.Errorf("trigger controller %s, want the player who LOST control %s", item.Controller, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	probe := effectstest.PushCatalogPermanent(g, me.ID, "Betrayer Probe", "Creature — Test", effectstest.LoseControlProbeOracle, false)
	g.ReadSnapshot(func() {})

	effectstest.StealForTest(t, g, probe, thief.ID)
	effectstest.PassPriorityAroundTable(t, g)
	thiefBefore := thief.Hand.Size()

	effectstest.AdvancePastCleanupForTest(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := thief.Hand.Size(); got != thiefBefore+2 {
		t.Errorf("the thief has %d cards after the theft expired, want %d — losing control at a "+
			"duration's end is still losing control", got, thiefBefore+2)
	}
}

// TestWhenYouGainControlOfThisFiresForTheNewController — Risky Move's
// half. The gaining player already controls the permanent by the time
// the event lands, so the ordinary item is theirs.
func TestWhenYouGainControlOfThisFiresForTheNewController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	probe := effectstest.PushCatalogPermanent(g, me.ID, "Turncoat Probe", "Creature — Test", effectstest.GainControlProbeOracle, false)
	g.ReadSnapshot(func() {})
	before := thief.Life

	effectstest.StealForTest(t, g, probe, thief.ID)

	item := effectstest.PendingTriggerFor(g, probe)
	if item == nil {
		t.Fatal("no trigger queued after the permanent changed control")
	}
	if item.Controller != thief.ID {
		t.Errorf("trigger controller %s, want the player who GAINED control %s", item.Controller, thief.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if thief.Life != before+3 {
		t.Errorf("the new controller is at %d life, want %d", thief.Life, before+3)
	}
}

// TestWheneverAnOpponentGainsControlOfAPermanentYouOwn — the Zedruu
// shape: a watcher permanent, a different permanent that moved, and
// OWNERSHIP as the link (CR 108.3).
func TestWheneverAnOpponentGainsControlOfAPermanentYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Donor Probe", "Creature — Test", effectstest.OwnedControlProbeOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, thief.ID, "Bear", "Creature — Bear", "", false)
	g.ReadSnapshot(func() {})
	before := me.Life

	// A permanent I own, taken by an opponent: the trigger.
	effectstest.StealForTest(t, g, mine, thief.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("life %d after an opponent took a permanent I own, want %d", me.Life, before+1)
	}

	// A permanent the opponent owns, taken by ME: not my trigger, and
	// not an opponent gaining anything.
	effectstest.StealForTest(t, g, theirs, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after I took a permanent an opponent owns, want %d — the clause reads "+
			"ownership and names an opponent as the gaining player", me.Life, before+1)
	}
}
