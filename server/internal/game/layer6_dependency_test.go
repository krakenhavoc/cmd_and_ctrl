package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer6_dependency_test.go is the engine-side suite for ADR 0067's
// amendment of 2026-10-23 (#2663): in the layer-6 bucket, a static that
// declares ReadsAbilities depends on every effect that would change what
// it applies to (CR 613.8a), waits for them whatever the timestamps
// (CR 613.8b), is asked again after each application (CR 613.8c), and a
// loop falls back to timestamp order (CR 613.8b's last sentence).
//
// Stubbed catalog hooks, like layer_dependency_test.go, so the rules are
// pinned apart from any one card. The catalog side — Sephara, Sky's
// Blade and Kwende, Pride of Femeref — is in
// cards/effects/keyword_reading_grant_test.go.

const (
	l6dReaderOracle  = "layer6dep-reader"
	l6dGranterOracle = "layer6dep-granter"
	l6dLoserOracle   = "layer6dep-loser"
	l6dCantOracle    = "layer6dep-cant-have"
	l6dLoopAOracle   = "layer6dep-loop-a"
	l6dLoopBOracle   = "layer6dep-loop-b"
	l6dWaiterOracle  = "layer6dep-waiter"
	l6dShroudsOracle = "layer6dep-flying-shrouds"
	l6dGroundOracle  = "layer6dep-shroud-grounds"
)

// l6dGrant is a layer-6 static that appends kw to the subject when the
// predicate holds.
func l6dGrant(subject *uuid.UUID, reads bool, when func(*Card) bool, kw string) StaticAbility {
	return StaticAbility{
		Layer:          Layer6Ability,
		ReadsAbilities: reads,
		AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
			return target.InstanceID == *subject && when(target)
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = AppendKeywordAbility(c.Abilities, kw)
		},
	}
}

// l6dLose is a layer-6 "the subject loses kw" (Gravity Sphere's shape:
// not a CR 613.1f "loses all abilities", so no ADR 0093 pre-sort).
func l6dLose(subject *uuid.UUID, reads bool, when func(*Card) bool, kw string) StaticAbility {
	return StaticAbility{
		Layer:          Layer6Ability,
		ReadsAbilities: reads,
		AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
			return target.InstanceID == *subject && when(target)
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			kept := c.Abilities[:0:0]
			for _, a := range c.Abilities {
				if a != kw {
					kept = append(kept, a)
				}
			}
			c.Abilities = kept
		},
	}
}

func l6dAlways(*Card) bool { return true }

func l6dHas(kw string) func(*Card) bool {
	return func(c *Card) bool { return HasKeywordItCanHave(c, kw) }
}

func l6dLacks(kw string) func(*Card) bool {
	return func(c *Card) bool { return !HasKeyword(c, kw) }
}

// l6dBoard puts the subject (its id written through `subject`), then
// the named enchantments in the order given (so in timestamp order), on
// the battlefield.
func l6dBoard(t *testing.T, subject *uuid.UUID, subjectKeywords []string, statics map[string][]StaticAbility, order ...string) *Game {
	t.Helper()
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility { return statics[oracleID] })
	seat := g.Seats[0].ID
	*subject = pushTypedTestCard(g, Card{
		Name: "Subject", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: subjectKeywords, Owner: seat, Controller: seat,
	})
	for _, oracle := range order {
		pushTypedTestCard(g, Card{
			Name: oracle, TypeLine: "Enchantment", OracleID: oracle,
			Owner: seat, Controller: seat,
		})
	}
	return g
}

// The reader is older than the effect that gives the subject flying: in
// timestamp order it would find no flying. It waits, whichever is older.
func TestAReaderWaitsForAGrantOfWhatItReads(t *testing.T) {
	for _, readerFirst := range []bool{true, false} {
		name := "granter older"
		order := []string{l6dGranterOracle, l6dReaderOracle}
		if readerFirst {
			name = "reader older"
			order = []string{l6dReaderOracle, l6dGranterOracle}
		}
		t.Run(name, func(t *testing.T) {
			sub := new(uuid.UUID)
			g := l6dBoard(t, sub, nil, map[string][]StaticAbility{
				l6dReaderOracle:  {l6dGrant(sub, true, l6dHas("flying"), "indestructible")},
				l6dGranterOracle: {l6dGrant(sub, false, l6dAlways, "flying")},
			}, order...)
			if !hasKeywordOnBattlefield(t, g, *sub, "flying") {
				t.Fatal("the subject has no flying")
			}
			if !hasKeywordOnBattlefield(t, g, *sub, "indestructible") {
				t.Error("the reader did not wait for the flying grant (CR 613.8b)")
			}
		})
	}
}

// The reader is older than the effect that takes the subject's printed
// flying away: in timestamp order it would grant indestructible and keep
// it, stronger than printed.
func TestAReaderWaitsForARemovalOfWhatItReads(t *testing.T) {
	for _, readerFirst := range []bool{true, false} {
		name := "loser older"
		order := []string{l6dLoserOracle, l6dReaderOracle}
		if readerFirst {
			name = "reader older"
			order = []string{l6dReaderOracle, l6dLoserOracle}
		}
		t.Run(name, func(t *testing.T) {
			sub := new(uuid.UUID)
			g := l6dBoard(t, sub, []string{"flying"}, map[string][]StaticAbility{
				l6dReaderOracle: {l6dGrant(sub, true, l6dHas("flying"), "indestructible")},
				l6dLoserOracle:  {l6dLose(sub, false, l6dAlways, "flying")},
			}, order...)
			if hasKeywordOnBattlefield(t, g, *sub, "flying") {
				t.Fatal("the subject kept flying")
			}
			if hasKeywordOnBattlefield(t, g, *sub, "indestructible") {
				t.Error("the reader granted indestructible to a creature that lost flying (CR 613.8b)")
			}
		})
	}
}

// "Can't have flying" is stripped after the bucket. The reader reads the
// record, so it depends on the lockout as it does on a "loses", and a
// flying grant newer than both does not bring indestructible.
func TestAReaderWaitsForACantHaveOfWhatItReads(t *testing.T) {
	sub := new(uuid.UUID)
	g := l6dBoard(t, sub, []string{"flying"}, map[string][]StaticAbility{
		l6dReaderOracle:  {l6dGrant(sub, true, l6dHas("flying"), "indestructible")},
		l6dCantOracle:    {{Layer: Layer6Ability, AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.InstanceID == *sub }, Apply: CantHaveKeywords("flying")}},
		l6dGranterOracle: {l6dGrant(sub, false, l6dAlways, "flying")},
	}, l6dReaderOracle, l6dCantOracle, l6dGranterOracle)
	if hasKeywordOnBattlefield(t, g, *sub, "flying") {
		t.Fatal("the subject has flying it can't have")
	}
	if hasKeywordOnBattlefield(t, g, *sub, "indestructible") {
		t.Error("the reader granted indestructible for flying the subject can't have")
	}
}

// One reader waiting on another, which is not a loop: "creatures with
// flying have shroud" (older) and "creatures with shroud lose flying"
// (newer — a reader that removes). Applying the first makes the second
// apply; applying the second, which applies to nothing yet, changes
// nothing. So the second waits for the first, the subject gets shroud and
// then loses flying, whichever entered first.
func TestAReaderWaitsForAnotherReader(t *testing.T) {
	for _, shroudsFirst := range []bool{true, false} {
		name := "grounder older"
		order := []string{l6dGroundOracle, l6dShroudsOracle}
		if shroudsFirst {
			name = "shrouder older"
			order = []string{l6dShroudsOracle, l6dGroundOracle}
		}
		t.Run(name, func(t *testing.T) {
			sub := new(uuid.UUID)
			g := l6dBoard(t, sub, []string{"flying"}, map[string][]StaticAbility{
				l6dShroudsOracle: {l6dGrant(sub, true, l6dHas("flying"), "shroud")},
				l6dGroundOracle:  {l6dLose(sub, true, l6dHas("shroud"), "flying")},
			}, order...)
			if !hasKeywordOnBattlefield(t, g, *sub, "shroud") || hasKeywordOnBattlefield(t, g, *sub, "flying") {
				t.Errorf("abilities %v, want shroud and no flying", layeredBattlefieldCard(t, g, *sub).Effective().Abilities)
			}
		})
	}
}

// A loop (CR 613.8b): "if it lacks shroud, it has hexproof" and "if it
// lacks hexproof, it has shroud". Each one, applied, takes the other's
// applicability away, so each depends on the other, the rule is ignored
// and the older applies first. The pass terminates with the timestamp
// answer.
func TestALayerSixDependencyLoopFallsBackToTimestampOrder(t *testing.T) {
	for _, aFirst := range []bool{true, false} {
		want, wantNot := "hexproof", "shroud"
		order := []string{l6dLoopAOracle, l6dLoopBOracle}
		if !aFirst {
			want, wantNot = "shroud", "hexproof"
			order = []string{l6dLoopBOracle, l6dLoopAOracle}
		}
		t.Run(want+" older", func(t *testing.T) {
			sub := new(uuid.UUID)
			g := l6dBoard(t, sub, nil, map[string][]StaticAbility{
				l6dLoopAOracle: {l6dGrant(sub, true, l6dLacks("shroud"), "hexproof")},
				l6dLoopBOracle: {l6dGrant(sub, true, l6dLacks("hexproof"), "shroud")},
			}, order...)
			if !hasKeywordOnBattlefield(t, g, *sub, want) || hasKeywordOnBattlefield(t, g, *sub, wantNot) {
				t.Errorf("abilities %v, want the older effect's %s only (CR 613.8b loop)",
					layeredBattlefieldCard(t, g, *sub).Effective().Abilities, want)
			}
		})
	}
}

// An effect that waits on a loop is not in it. "Creatures with shroud
// have indestructible" is older than both halves of the loop, and depends
// on the half that grants shroud. The loop is applied first, by timestamp
// (shroud's half is the older), and the waiter then finds shroud. Falling
// back to the oldest remaining effect instead would apply the waiter
// first, to nothing.
func TestAnEffectWaitingOnALoopAppliesAfterIt(t *testing.T) {
	sub := new(uuid.UUID)
	g := l6dBoard(t, sub, nil, map[string][]StaticAbility{
		l6dWaiterOracle: {l6dGrant(sub, true, l6dHas("shroud"), "indestructible")},
		l6dLoopBOracle:  {l6dGrant(sub, true, l6dLacks("hexproof"), "shroud")},
		l6dLoopAOracle:  {l6dGrant(sub, true, l6dLacks("shroud"), "hexproof")},
	}, l6dWaiterOracle, l6dLoopBOracle, l6dLoopAOracle)
	got := layeredBattlefieldCard(t, g, *sub).Effective().Abilities
	if !hasKeywordOnBattlefield(t, g, *sub, "shroud") || hasKeywordOnBattlefield(t, g, *sub, "hexproof") {
		t.Fatalf("abilities %v, want the older loop half's shroud only", got)
	}
	if !hasKeywordOnBattlefield(t, g, *sub, "indestructible") {
		t.Errorf("abilities %v: the effect waiting on the loop applied before it", got)
	}
}

// A static that does not declare ReadsAbilities keeps the timestamp
// order it always had: the declaration is what turns the trial on.
func TestAnUndeclaredReaderKeepsTimestampOrder(t *testing.T) {
	sub := new(uuid.UUID)
	g := l6dBoard(t, sub, nil, map[string][]StaticAbility{
		l6dReaderOracle:  {l6dGrant(sub, false, l6dHas("flying"), "indestructible")},
		l6dGranterOracle: {l6dGrant(sub, false, l6dAlways, "flying")},
	}, l6dReaderOracle, l6dGranterOracle)
	if hasKeywordOnBattlefield(t, g, *sub, "indestructible") {
		t.Error("an undeclared reader was reordered")
	}
}

func TestDependsOnItself(t *testing.T) {
	// 0 -> 1 -> 2 -> 1: 1 and 2 are in a loop, 0 only waits on it.
	edges := [][]bool{
		{false, true, false},
		{false, false, true},
		{false, true, false},
	}
	for i, want := range []bool{false, true, true} {
		if got := dependsOnItself(edges, i); got != want {
			t.Errorf("dependsOnItself(%d) = %v, want %v", i, got, want)
		}
	}
}
