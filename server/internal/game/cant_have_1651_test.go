package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// cant_have_1651_test.go — #1651, the engine half of ADR 0038's
// amendment of 2026-09-28: "can't have" beats a grant of any timestamp
// (B2), and a hexproof waiver can be a turn-scoped data record (B1). The
// catalog half (Detection Tower, Arcane Lighthouse, the Archetypes) is in
// cards/effects/hexproof_1651_test.go.

func abilitiesOf(t *testing.T, g *Game, id uuid.UUID) []string {
	t.Helper()
	return scopedEffectChar(t, g, id).Abilities
}

// The heart of B2: a can't-have recorded at timestamp T beats a grant at
// T+1. A plain removal at T would not (CR 613.7), which the second half
// pins so the difference is visible in one test.
func TestCantHaveBeatsALaterGrant(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushScopedTestCreature(g, me.ID, 2, 2)

	registerScopedEffectForTest(t, g, id, []Mod{CantHaveKeywordsMod("hexproof")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{AddKeywordsMod("hexproof", "trample")}, IndefiniteDuration())

	got := abilitiesOf(t, g, id)
	if hasAbility(got, "hexproof") {
		t.Errorf("abilities = %v: a grant newer than the can't-have must not stick", got)
	}
	if !hasAbility(got, "trample") {
		t.Errorf("abilities = %v: only the named keyword is refused", got)
	}

	plain := pushScopedTestCreature(g, me.ID, 2, 2)
	registerScopedEffectForTest(t, g, plain, []Mod{RemoveKeywordsMod("hexproof")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, plain, []Mod{AddKeywordsMod("hexproof")}, IndefiniteDuration())
	if !hasAbility(abilitiesOf(t, g, plain), "hexproof") {
		t.Error("control: a plain removal is timestamp-ordered, so the later grant wins")
	}
}

// "Lose" needs no separate removal: the strip takes the printed
// keyword too, and the object stops being hexproof at the targeting
// choke point.
func TestCantHaveStripsThePrintedKeyword(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushProtectedCreature(g, opp, "Their Bogle", "hexproof", "shroud")
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: theirs, OldZone: ZoneHand, NewZone: ZoneBattlefield})
	})
	if legalCreatureTargets(g, me.ID)[theirs] {
		t.Fatal("baseline: a hexproof, shrouded creature is no target")
	}
	registerScopedEffectForTest(t, g, theirs, []Mod{CantHaveKeywordsMod("hexproof", "shroud")}, IndefiniteDuration())
	if got := abilitiesOf(t, g, theirs); hasAbility(got, "hexproof") || hasAbility(got, "shroud") {
		t.Errorf("abilities = %v, want hexproof and shroud stripped", got)
	}
	if !legalCreatureTargets(g, me.ID)[theirs] {
		t.Error("with neither keyword the creature is a legal target")
	}
}

// Nothing clears CantHave, a CR 613.1f "loses all abilities" included:
// the can't-have belongs to the effect's source, like a restriction.
func TestCantHaveSurvivesALaterLoseAllAbilities(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{CantHaveKeywordsMod("flying")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllAbilitiesMod("flying")}, IndefiniteDuration())

	c := scopedEffectChar(t, g, id)
	if hasAbility(c.Abilities, "flying") {
		t.Errorf("abilities = %v: the removal's own re-grant must not beat the can't-have", c.Abilities)
	}
	if !hasAbility(c.CantHave, "flying") {
		t.Errorf("CantHave = %v: a later ability removal must not clear it", c.CantHave)
	}
}

// A can't-have record is data: it survives capture → JSON → restore
// and still beats a grant registered in the restored game.
func TestCantHaveIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{CantHaveKeywordsMod("hexproof")}, g.UntilEndOfTurnDuration())

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	registerScopedEffectForTest(t, restored, id, []Mod{AddKeywordsMod("hexproof")}, IndefiniteDuration())
	if hasAbility(abilitiesOf(t, restored, id), "hexproof") {
		t.Error("the restored can't-have must still beat a later grant")
	}
}

func TestCantHaveModWithoutKeywordsPanics(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	defer func() {
		if recover() == nil {
			t.Error("a cantHaveKeywords mod naming nothing registered without a panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
			[]Mod{{Kind: ModCantHaveKeywords}}, IndefiniteDuration(), "test")
	})
}

// registerWaiver installs Detection Tower's record for `controller`.
func registerWaiver(t *testing.T, g *Game, controller uuid.UUID) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeOpponentsAndTheirCreatures, controller,
			[]Mod{WaiveHexproofMod()}, g.UntilEndOfTurnDuration(), "test waiver")
	})
	if !ok {
		t.Fatal("the waiver registered nothing")
	}
}

func legalPlayerTargets(g *Game, by uuid.UUID) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, id := range g.LegalTargetsForEffect(SourceChooser(by), anyPlayerSpec()).Players {
			out[id] = true
		}
	})
	return out
}

// B1: the scoped waiver answers at the same choke point as the static
// one, for its controller only, over a live set, for creatures and
// players both — and only for hexproof.
func TestScopedHexproofWaiver(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pushProtectedCreature(g, opp, "Their Bogle", "hexproof")
	mine := pushProtectedCreature(g, me, "My Bogle", "hexproof")
	shrouded := pushProtectedCreature(g, opp, "Their Ledgewalker", "shroud")
	opp.Statics = append(opp.Statics, PlayerStatic{Keyword: KeywordHexproof})

	if legalCreatureTargets(g, me.ID)[theirs] || legalPlayerTargets(g, me.ID)[opp.ID] {
		t.Fatal("baseline: hexproof refuses an opponent")
	}
	registerWaiver(t, g, me.ID)

	if !legalCreatureTargets(g, me.ID)[theirs] {
		t.Error("the waiver's controller may target an opponent's hexproof creature")
	}
	if !legalPlayerTargets(g, me.ID)[opp.ID] {
		t.Error("the waiver's controller may target a hexproof opponent (CR 702.11d)")
	}
	if legalCreatureTargets(g, third.ID)[theirs] || legalPlayerTargets(g, third.ID)[opp.ID] {
		t.Error(`"spells and abilities you control": a third player gets nothing`)
	}
	if legalCreatureTargets(g, opp.ID)[mine] {
		t.Error("the controller's own creature is not an opponent's and keeps its hexproof")
	}
	if legalCreatureTargets(g, me.ID)[shrouded] {
		t.Error("shroud is not hexproof and is not waived")
	}
	// Live: a hexproof creature arriving after the waiver is covered.
	late := pushProtectedCreature(g, opp, "Late Bogle", "hexproof")
	if !legalCreatureTargets(g, me.ID)[late] {
		t.Error("the waiver's set is a live rule, so a later creature is covered")
	}

	advancePastScopedCleanup(t, g)
	// The test's player static is a turn-scoped grant too; put it back
	// so the player half is asked about the waiver alone.
	opp.Statics = append(opp.Statics, PlayerStatic{Keyword: KeywordHexproof})
	if legalCreatureTargets(g, me.ID)[theirs] || legalPlayerTargets(g, me.ID)[opp.ID] {
		t.Error("the waiver ends in cleanup")
	}
}

// Undo is Clone + RestoreFrom, and a restore point carries the record.
func TestScopedHexproofWaiverSurvivesUndoAndRestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushProtectedCreature(g, opp, "Their Bogle", "hexproof")
	before := g.Clone()
	registerWaiver(t, g, me.ID)
	after := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if legalCreatureTargets(g, me.ID)[theirs] {
		t.Error("undo past the activation takes the waiver away")
	}
	g.WithWriteLock(func() { g.RestoreFrom(after) })
	if !legalCreatureTargets(g, me.ID)[theirs] {
		t.Error("redo brings it back")
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if !legalCreatureTargets(restored, me.ID)[theirs] {
		t.Error("the restored game keeps the waiver")
	}
}

// Only the new scope names players, so no earlier record changes
// meaning.
func TestScopeCoversPlayerOnlyForTheNewScope(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	for _, s := range []AffectedScope{ScopeNone, ScopeOpponentsCreatures, ScopeGame, ScopeYourPermanents,
		ScopeYourCreatures, ScopeCreaturesWithoutFlying} {
		if scopeCoversPlayer(s, me, opp) {
			t.Errorf("scope %q must name no player", s)
		}
	}
	if !scopeCoversPlayer(ScopeOpponentsAndTheirCreatures, me, opp) || scopeCoversPlayer(ScopeOpponentsAndTheirCreatures, me, me) {
		t.Error("opponentsAndTheirCreatures names the controller's opponents and not the controller")
	}
}
