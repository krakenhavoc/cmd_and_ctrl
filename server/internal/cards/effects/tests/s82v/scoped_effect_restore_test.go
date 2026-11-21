package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMassDiminishSurvivesARestoreAndStillEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Bear")

	effectstest.CastCatalogSpell(t, g, "Mass Diminish", "Sorcery", effectstest.MassDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if got := effectstest.EffectivePower(t, restored, theirs); got != 1 {
		t.Fatalf("restored: power %d, want 1", got)
	}
	for seat := 1; seat <= 3; seat++ {
		effectstest.AdvancePastCleanupForTest(t, restored)
		if got := effectstest.EffectivePower(t, restored, theirs); got != 1 {
			t.Fatalf("restored, after seat %d's turn: power %d, want 1 — the effect ended early", seat, got)
		}
	}
	effectstest.AdvancePastCleanupForTest(t, restored)
	if got := effectstest.EffectivePower(t, restored, theirs); got != 2 {
		t.Errorf("restored, after the caster's next turn began: power %d, want the printed 2", got)
	}
	if n := len(restored.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects outlived their duration in the restored game", n)
	}
}

func TestTreeOfPerditionToughnessSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = 37
	tree := effectstest.LxTree(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, tree, 0,
		game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if got := effectstest.EffectiveToughness(t, restored, tree); got != 37 {
		t.Errorf("restored Tree's toughness = %d, want 37 — the indefinite set did not survive", got)
	}
}

func TestEnduringCuriosityStaysAnEnchantmentAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cat := effectstest.PushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", effectstest.EnduringCuriosityOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cat) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cat) {
		t.Fatal("setup: Enduring Curiosity did not come back")
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
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

// TestAnEarthbentKyoshiLandIsARestorePoint is tier 2's payoff (#1497).
// Chapter II's Island and the earthbend animation became data in tier
// 1, but earthbend's "return it tapped when it dies or is exiled" was a
// closure-bearing delayed trigger, so the land still froze its table's
// restore point for as long as it lived. The return is a registered
// body and condition now: the table is a full restore point, and in the
// restored game the land still comes back tapped when it dies.
func TestAnEarthbentKyoshiLandIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	land := effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.ImportAndCast(t, g, effectstest.KyoshiRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("an earthbent land still blocks the restore point: %+v", snap.Continuations)
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	c := effectstest.FindBattlefieldCardByID(restored, land)
	if c == nil {
		t.Fatal("the land did not survive the restore")
	}
	if !c.HasSubtype("Island") {
		t.Error("restored: the land is no longer an Island")
	}
	if !c.IsCreature() {
		t.Error("restored: the earthbent land is no longer a creature")
	}

	// The return still works in the restored game.
	restored.WithWriteLock(func() { _ = restored.DestroyPermanentForEffect(land) })
	effectstest.PassPriorityAroundTable(t, restored)
	back := effectstest.FindBattlefieldCardByID(restored, land)
	if back == nil {
		t.Fatal("restored: the earthbent land died and did not come back")
	}
	if !back.Tapped {
		t.Error("restored: the land came back untapped; earthbend returns it tapped")
	}
}

// --- phasing (#1497 review) ------------------------------------------
//
// Phasing is not a zone change (CR 702.26d): a permanent keeps its
// instance and its entry stamp, and an indefinite effect on it applies
// again when it phases in. The pin that garbage-collects a record used
// to read the battlefield alone and dropped the record the moment the
// permanent phased out; these pin the fix, live and across a restore
// taken while the permanent is out.

// phaseOut phases one permanent out.
func phaseOut(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, id); err != nil {
			t.Fatalf("PhaseOutForEffect: %v", err)
		}
	})
	if !phasedOut(g, id) {
		t.Fatal("setup: the permanent did not phase out")
	}
}

func phasedOut(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.WithWriteLock(func() { out = g.IsPhasedOutForEffect(id) })
	return out
}

// untilPhasedIn advances turns until id has phased back in.
func untilPhasedIn(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && phasedOut(g, id); i++ {
		effectstest.AdvancePastCleanupForTest(t, g)
	}
	if phasedOut(g, id) {
		t.Fatal("the permanent never phased back in")
	}
}

func activatedTree(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = 37
	tree := effectstest.LxTree(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, tree, 0,
		game.ActivateAbilityParams{Targets: effectstest.AcPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, tree); got != 37 {
		t.Fatalf("setup: toughness %d, want 37", got)
	}
	return g, tree
}

func TestTreeOfPerditionToughnessSurvivesAPhaseCycle(t *testing.T) {
	g, tree := activatedTree(t)
	phaseOut(t, g, tree)
	untilPhasedIn(t, g, tree)
	if got := effectstest.EffectiveToughness(t, g, tree); got != 37 {
		t.Errorf("after phasing in: toughness %d, want 37 — phasing is not a zone change (CR 702.26d)", got)
	}
}

func TestTreeOfPerditionToughnessSurvivesARestoreWhilePhasedOut(t *testing.T) {
	g, tree := activatedTree(t)
	phaseOut(t, g, tree)
	restored := effectstest.RestoreRoundTrip(t, g, true)
	untilPhasedIn(t, restored, tree)
	if got := effectstest.EffectiveToughness(t, restored, tree); got != 37 {
		t.Errorf("restored, after phasing in: toughness %d, want 37", got)
	}
}

// kyoshiIslandLand plays chapter II onto a Forest and returns the game
// and the land.
func kyoshiIslandLand(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	land := effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.ImportAndCast(t, g, effectstest.KyoshiRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	return g, land
}

func assertIsland(t *testing.T, g *game.Game, land uuid.UUID, when string) {
	t.Helper()
	c := effectstest.FindBattlefieldCardByID(g, land)
	if c == nil {
		t.Fatalf("%s: the land is not on the battlefield", when)
	}
	if !c.HasSubtype("Island") {
		t.Errorf("%s: the land is no longer an Island (CR 702.26d)", when)
	}
	if !c.IsCreature() {
		t.Errorf("%s: the earthbent land is no longer a creature (CR 702.26d)", when)
	}
}

func TestKyoshiIslandSurvivesAPhaseCycle(t *testing.T) {
	g, land := kyoshiIslandLand(t)
	phaseOut(t, g, land)
	untilPhasedIn(t, g, land)
	assertIsland(t, g, land, "after phasing in")
}

func TestKyoshiIslandSurvivesARestoreWhilePhasedOut(t *testing.T) {
	g, land := kyoshiIslandLand(t)
	phaseOut(t, g, land)
	// Strict since tier 2: earthbend's return is data too.
	restored := effectstest.RestoreRoundTrip(t, g, true)
	untilPhasedIn(t, restored, land)
	assertIsland(t, restored, land, "restored, after phasing in")
}

// TestEnduringCuriosityStaysAnEnchantmentThroughAPhaseCycle: "It's an
// enchantment" states no duration, so it is indefinite on the returned
// object — not a "for as long as" duration that CR 702.26f would end
// when the permanent phases out.
func TestEnduringCuriosityStaysAnEnchantmentThroughAPhaseCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cat := effectstest.PushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", effectstest.EnduringCuriosityOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cat) })
	effectstest.PassPriorityAroundTable(t, g)
	phaseOut(t, g, cat)
	untilPhasedIn(t, g, cat)
	c := effectstest.FindBattlefieldCardByID(g, cat)
	if c == nil {
		t.Fatal("Enduring Curiosity is not on the battlefield")
	}
	if c.IsCreature() {
		t.Error("after phasing in it is a creature again — the effect ended on the phase-out")
	}
}
