package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_effect_restore_test.go pins what ADR 0041 phase 3's first
// slice (#1497) is for: a card whose continuous effect used to be a
// closure-bearing ScopedStatic — and so froze its table's restore
// point for as long as the effect lived — now leaves a game that IS a
// restore point, and the effect is still there, and still ends, in the
// restored game.
//
// Every test asserts what a player can see (P/T, types, subtypes)
// across a real capture → JSON → restore, not merely that a record was
// written.

// restoreRoundTrip captures g, sends it through JSON and restores it.
// strict demands a full restore point (RestoreStrict, the boot path);
// otherwise Restore is used and the census is only reported.
func restoreRoundTrip(t *testing.T, g *game.Game, strict bool) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if strict && !snap.Restorable() {
		t.Fatalf("the game is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var restored *game.Game
	if strict {
		restored, err = decoded.RestoreStrict()
	} else {
		restored, err = decoded.Restore()
	}
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	restored.WithWriteLock(func() { restored.RecomputeLayersIfStaleLocked() })
	return restored
}

func TestMassDiminishSurvivesARestoreAndStillEnds(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	theirs := ctrlPushCreature(g, opp.ID, "Bear")

	castCatalogSpell(t, g, "Mass Diminish", "Sorcery", massDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	restored := restoreRoundTrip(t, g, true)
	if got := effectivePower(t, restored, theirs); got != 1 {
		t.Fatalf("restored: power %d, want 1", got)
	}
	for seat := 1; seat <= 3; seat++ {
		advancePastCleanupForTest(t, restored)
		if got := effectivePower(t, restored, theirs); got != 1 {
			t.Fatalf("restored, after seat %d's turn: power %d, want 1 — the effect ended early", seat, got)
		}
	}
	advancePastCleanupForTest(t, restored)
	if got := effectivePower(t, restored, theirs); got != 2 {
		t.Errorf("restored, after the caster's next turn began: power %d, want the printed 2", got)
	}
	if n := len(restored.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects outlived their duration in the restored game", n)
	}
}

func TestTreeOfPerditionToughnessSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = 37
	tree := lxTree(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, tree, 0,
		game.ActivateAbilityParams{Targets: acPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	restored := restoreRoundTrip(t, g, true)
	if got := effectiveToughness(t, restored, tree); got != 37 {
		t.Errorf("restored Tree's toughness = %d, want 37 — the indefinite set did not survive", got)
	}
}

func TestEnduringCuriosityStaysAnEnchantmentAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cat := pushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cat) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cat) {
		t.Fatal("setup: Enduring Curiosity did not come back")
	}

	restored := restoreRoundTrip(t, g, true)
	var isCreature, isEnchantment, found bool
	restored.ReadSnapshot(func() {
		for _, c := range restored.Battlefield.Cards {
			if c.InstanceID == cat {
				found = true
				isCreature, isEnchantment = c.IsCreature(), c.IsEnchantment()
			}
		}
	})
	if !found {
		t.Fatal("Enduring Curiosity did not survive the restore")
	}
	if isCreature {
		t.Error("restored: it is a creature again — the 'it's not a creature' effect was lost")
	}
	if !isEnchantment {
		t.Error("restored: it stopped being an enchantment")
	}
}

// TestKyoshiIslandIsDataButTheEarthbendReturnStillBlocks is the honest
// half of the slice. Chapter II's Island and the earthbend animation
// are data now, but earthbend also schedules its "return it tapped"
// delayed trigger, which stays a closure until ADR 0041 phase 3's
// tier 2. So the census must name exactly that — and nothing about a
// scoped static — and a (non-strict) restore must keep the Island.
func TestKyoshiIslandIsDataButTheEarthbendReturnStillBlocks(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	importAndCast(t, g, kyoshiRow(), me)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	pickCard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)

	snap := g.CaptureSnapshot()
	if snap.Continuations.ScopedStatics != 0 {
		t.Errorf("census counts %d scoped statics; the Island and the animation should be data", snap.Continuations.ScopedStatics)
	}
	if snap.Continuations.DelayedTriggerEffects != 1 {
		t.Errorf("census counts %d delayed triggers, want 1 (earthbend's return, tier 2)", snap.Continuations.DelayedTriggerEffects)
	}
	for _, l := range snap.Continuations.Labels {
		if strings.Contains(l, "scoped static") {
			t.Errorf("census label names a scoped static: %q", l)
		}
	}

	restored := restoreRoundTrip(t, g, false)
	c := findBattlefieldCardByID(restored, findBattlefieldByName(restored, "Forest"))
	if c == nil {
		t.Fatal("the land did not survive the restore")
	}
	if !c.HasSubtype("Island") {
		t.Error("restored: the land is no longer an Island")
	}
	if !c.IsCreature() {
		t.Error("restored: the earthbent land is no longer a creature")
	}
}
