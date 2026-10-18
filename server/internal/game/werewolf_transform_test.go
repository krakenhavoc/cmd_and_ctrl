package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// werewolf_transform_test.go — the two seams the Innistrad werewolves
// needed (ADR 0132 amendment 2026-10-16, #4071): the previous turn's
// spells counted across every player, and "can't transform" (Immerwolf).
// The cards themselves are pinned in cards/effects.

// TestLastTurnSpellConditions drives the real rotation on a four-player
// table: what every player cast during seat 0's turn decides the two
// werewolf conditions for the turn after it.
func TestLastTurnSpellConditions(t *testing.T) {
	for _, tc := range []struct {
		name     string
		casts    map[int]int // seat -> spells cast during seat 0's turn
		noSpells bool        // "if no spells were cast last turn"
		twoByOne bool        // "if a player cast two or more spells last turn"
	}{
		{"no spells", nil, true, false},
		{"one spell by the active player", map[int]int{0: 1}, false, false},
		{"one spell by an opponent", map[int]int{2: 1}, false, false},
		{"one player cast two", map[int]int{1: 2}, false, true},
		{"one player cast three", map[int]int{0: 3}, false, true},
		{"two players cast one each", map[int]int{0: 1, 3: 1}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			if g.NoSpellsWereCastLastTurnForEffect() || g.APlayerCastTwoOrMoreSpellsLastTurnForEffect() {
				t.Fatal("the game's first turn has no last turn, so neither condition can be true")
			}
			g.WithWriteLock(func() {
				g.SpellsCastThisTurn = map[uuid.UUID]CastTally{}
				for seat, n := range tc.casts {
					g.SpellsCastThisTurn[g.Seats[seat].ID] = CastTally{Total: n}
				}
			})
			advanceOneTurn(t, g)
			if got := g.NoSpellsWereCastLastTurnForEffect(); got != tc.noSpells {
				t.Errorf("no spells were cast last turn = %v, want %v", got, tc.noSpells)
			}
			if got := g.APlayerCastTwoOrMoreSpellsLastTurnForEffect(); got != tc.twoByOne {
				t.Errorf("a player cast two or more spells last turn = %v, want %v", got, tc.twoByOne)
			}
		})
	}
}

// TestLastTurnSpellRecordIsTheTurnBeforeOnly: spells cast during the
// current turn do not move the record; it is replaced only as the turn
// ends.
func TestLastTurnSpellRecordIsTheTurnBeforeOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceOneTurn(t, g) // seat 0 cast nothing
	if !g.NoSpellsWereCastLastTurnForEffect() {
		t.Fatal("a quiet turn did not record no spells")
	}
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{g.Seats[g.Turn.ActiveSeat].ID: {Total: 2}}
	})
	if !g.NoSpellsWereCastLastTurnForEffect() || g.APlayerCastTwoOrMoreSpellsLastTurnForEffect() {
		t.Error("this turn's spells changed the record of last turn")
	}
	advanceOneTurn(t, g)
	if g.NoSpellsWereCastLastTurnForEffect() || !g.APlayerCastTwoOrMoreSpellsLastTurnForEffect() {
		t.Error("the next turn does not read the two spells as last turn's")
	}
}

// TestLastTurnSpellRecordRoundTripsTheSnapshot: the record survives a
// JSON snapshot round trip and an undo, and a file written before the
// record existed (the day/night fields only) reads as no record.
func TestLastTurnSpellRecordRoundTripsTheSnapshot(t *testing.T) {
	g := newActiveGame(t)
	want := DayNightState{Designation: DesignationDay, PrevTurnSpells: 1, PrevTurnKnown: true,
		PrevTurnAllSpells: 3, PrevTurnMostSpells: 2, PrevTurnAllKnown: true}
	g.WithWriteLock(func() { g.DayNight = want })

	raw, err := json.Marshal(g.CaptureSnapshot())
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
	if restored.DayNight != want {
		t.Errorf("restored %+v, want %+v", restored.DayNight, want)
	}
	if !restored.APlayerCastTwoOrMoreSpellsLastTurnForEffect() {
		t.Error("the restored game lost the two-spell record")
	}

	snap := g.Clone()
	g.WithWriteLock(func() { g.DayNight = DayNightState{} })
	g.RestoreFrom(snap)
	if g.DayNight != want {
		t.Errorf("undo restored %+v, want %+v", g.DayNight, want)
	}

	// An older file: the day/night check's record and nothing else.
	legacy := DayNightState{Designation: DesignationDay, PrevTurnKnown: true}
	decoded.DayNight = &legacy
	old, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict (legacy): %v", err)
	}
	if old.NoSpellsWereCastLastTurnForEffect() || old.APlayerCastTwoOrMoreSpellsLastTurnForEffect() {
		t.Error("a file with no all-player record read as one (it would flip every werewolf on the board)")
	}
}

// --- can't transform (Immerwolf) ---------------------------------------

const cantTransformRestrictorOracle = "immerwolf-test-restrictor"

// withNonHumanWerewolfLock installs Immerwolf's third ability on the
// restrictor fixture: "Non-Human Werewolves you control can't transform".
func withNonHumanWerewolfLock(t *testing.T) {
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != cantTransformRestrictorOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.Controller == source.Controller && target.HasSubtype("Werewolf") && !target.HasSubtype("Human")
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.CantTransform = true },
		}}
	})
}

func pushTransformRestrictor(g *Game, controller uuid.UUID) uuid.UUID {
	return pushTypedTestCard(g, Card{
		OracleID: cantTransformRestrictorOracle, Name: "Test Immerwolf",
		TypeLine: "Creature — Wolf", Power: 2, Toughness: 2,
		Owner: controller, Controller: controller,
	})
}

// oldWerewolfFixture is an Innistrad werewolf: a transforming
// double-faced card with no daybound or nightbound, a Human Werewolf
// front and a Werewolf back.
func oldWerewolfFixture(owner uuid.UUID) Card {
	c := werewolfFixture(owner)
	c.OracleID = "bbbbbbbb-bbbb-cccc-dddd-eeeeeeeeeeee"
	c.Faces[0].Keywords = nil
	c.Faces[1].Keywords = nil
	c.SetFace(0)
	c.Keywords = nil // SetFace keeps the card's list when the face has none
	return c
}

func transformForTest(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.TransformPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("TransformPermanentForEffect: %v (an instruction that does nothing is not an error)", err)
	}
}

// TestCantTransformStopsANonHumanWerewolfOnly: the Human Werewolf front
// is not covered, so it transforms; the back it lands on is, so it can't
// transform back, and CanTransformForEffect says so. Another player's
// werewolf is not "you control".
func TestCantTransformStopsANonHumanWerewolfOnly(t *testing.T) {
	withNonHumanWerewolfLock(t)
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	pushTransformRestrictor(g, me)
	mine := pushTypedTestCard(g, oldWerewolfFixture(me))
	theirs := pushTypedTestCard(g, oldWerewolfFixture(them))

	transformForTest(t, g, mine)
	if c := findBattlefieldCard(g, mine); c.ActiveFace != 1 {
		t.Fatalf("a Human Werewolf front did not transform (ActiveFace %d): it is not a non-Human Werewolf", c.ActiveFace)
	}
	var can bool
	g.WithWriteLock(func() { can = g.CanTransformForEffect(mine) })
	if can {
		t.Error("CanTransformForEffect says the locked back face can transform")
	}
	before := dnCount(g, EventTransform)
	transformForTest(t, g, mine)
	if c := findBattlefieldCard(g, mine); c.ActiveFace != 1 {
		t.Fatalf("a non-Human Werewolf you control transformed back (ActiveFace %d)", c.ActiveFace)
	}
	if got := dnCount(g, EventTransform) - before; got != 0 {
		t.Errorf("a refused transform emitted %d EventTransform", got)
	}
	var err error
	g.WithWriteLock(func() { err = g.ConvertPermanentForEffect(mine) })
	if err != nil || findBattlefieldCard(g, mine).ActiveFace != 1 {
		t.Errorf("a permanent that can't transform converted (CR 701.28f): err=%v", err)
	}

	transformForTest(t, g, theirs)
	transformForTest(t, g, theirs)
	if c := findBattlefieldCard(g, theirs); c.ActiveFace != 0 {
		t.Errorf("an opponent's werewolf was held by your restriction (ActiveFace %d)", c.ActiveFace)
	}
}

// TestCantTransformHoldsANightboundWerewolfByDay: the day turn-over is
// a transform too, so a nightbound Werewolf you control stays on its
// back face while the restriction lasts (CR 101.2), and turns over the
// moment the restrictor leaves (CR 702.145f's "any time").
func TestCantTransformHoldsANightboundWerewolfByDay(t *testing.T) {
	withNonHumanWerewolfLock(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := enterFromHand(t, g, werewolfFixture(me)) // day: front up
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if c := findBattlefieldCard(g, id); c.ActiveFace != 1 {
		t.Fatalf("night did not turn the Human Werewolf front over (ActiveFace %d)", c.ActiveFace)
	}
	wolf := pushTransformRestrictor(g, me)
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if c := findBattlefieldCard(g, id); c.ActiveFace != 1 {
		t.Fatalf("day turned a nightbound non-Human Werewolf over under the restriction (ActiveFace %d)", c.ActiveFace)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wolf); err != nil {
			t.Fatalf("destroy the restrictor: %v", err)
		}
	})
	if c := findBattlefieldCard(g, id); c.ActiveFace != 0 {
		t.Errorf("the restriction ended by day and the nightbound werewolf stayed up (ActiveFace %d)", c.ActiveFace)
	}
	AssertFaceInvariant(t, g)
}

// TestCantTransformDoesNotStopAHumanWerewolfFlippingToNight: the
// daybound front is a Human Werewolf, which the restriction does not
// name, so night still turns it over.
func TestCantTransformDoesNotStopAHumanWerewolfFlippingToNight(t *testing.T) {
	withNonHumanWerewolfLock(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	pushTransformRestrictor(g, me)
	id := enterFromHand(t, g, werewolfFixture(me))
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if c := findBattlefieldCard(g, id); c.ActiveFace != 1 {
		t.Errorf("night did not turn a Human Werewolf over under a non-Human restriction (ActiveFace %d)", c.ActiveFace)
	}
}
