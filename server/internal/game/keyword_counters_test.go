package game

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"
)

// keyword_counters_test.go — ADR 0101 (#1753), CR 122.1b: the engine
// reads keyword counters itself. Every rule the ADR states has a test
// here; the catalog half (Perennation, the five cards that used to
// carry b24KeywordCounterGrant) is in cards/effects.

// cr1221bKeywords is CR 122.1b's list, verbatim, in the pinned edition:
// "flying, first strike, double strike, deathtouch, decayed, exalted,
// haste, hexproof, indestructible, lifelink, menace, reach, shadow,
// trample, and vigilance".
var cr1221bKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "decayed",
	"exalted", "haste", "hexproof", "indestructible", "lifelink", "menace",
	"reach", "shadow", "trample", "vigilance",
}

// A kind outside CR 122.1b's list can never join the table.
func TestKeywordCounterKindsAreCR1221b(t *testing.T) {
	allowed := map[string]bool{}
	for _, kw := range cr1221bKeywords {
		allowed[kw] = true
	}
	for _, kind := range KeywordCounterKinds() {
		if !allowed[kind] {
			t.Errorf("keyword counter kind %q is not in CR 122.1b's list", kind)
		}
	}
}

// Every kind is a keyword the engine enforces (ADR 0014's closedness
// rule), and every CR 122.1b keyword the engine enforces IS a kind —
// so a keyword that joins canonicalKeywords without its counter fails
// here. Exalted joined with #2538 and decayed, the last, with #2650.
func TestKeywordCounterKindsAreCanonicalKeywords(t *testing.T) {
	for _, kind := range KeywordCounterKinds() {
		if !canonicalKeywords[kind] {
			t.Errorf("keyword counter kind %q is not a canonical keyword", kind)
		}
	}
	for _, kw := range cr1221bKeywords {
		if canonicalKeywords[kw] && !IsKeywordCounter(kw) {
			t.Errorf("%q is a canonical keyword and a CR 122.1b keyword counter, but not in keywordCounterKinds", kw)
		}
	}
	for _, out := range []string{"hexproof from white", "Flying", "+1/+1", "shield", "stun", "protection from red"} {
		if IsKeywordCounter(out) {
			t.Errorf("IsKeywordCounter(%q) = true, want false", out)
		}
	}
}

// steppedClock makes every engine timestamp distinct and increasing, so
// "earlier" and "later" in a test mean what the test says.
func steppedClock(t *testing.T) {
	t.Helper()
	now := int64(1_000_000)
	restore := SetClockForTest(func() int64 { now += 10; return now })
	t.Cleanup(restore)
}

func addCounterForTest(t *testing.T, g *Game, id uuid.UUID, kind string, delta int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, kind, delta) })
	if err != nil {
		t.Fatalf("AddCounterForEffect(%q, %d): %v", kind, delta, err)
	}
}

func countOf(list []string, want string) int {
	n := 0
	for _, a := range list {
		if a == want {
			n++
		}
	}
	return n
}

func hasKeywordNow(t *testing.T, g *Game, id uuid.UUID, kw string) bool {
	t.Helper()
	ch := scopedEffectChar(t, g, id)
	return hasAbility(ch.Abilities, kw)
}

func TestAKeywordCounterGivesItsKeyword(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	if hasKeywordNow(t, g, id, "flying") {
		t.Fatal("baseline: a vanilla Bear does not fly")
	}
	addCounterForTest(t, g, id, CounterFlying, 1)
	if !hasKeywordNow(t, g, id, "flying") {
		t.Fatal("a flying counter must give flying (CR 122.1b)")
	}
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		if !HasKeyword(c, "flying") {
			t.Error("HasKeyword must read the counter's keyword off the effective list")
		}
	})
	// A counter kind that is not a keyword counter grants nothing.
	addCounterForTest(t, g, id, "Trample", 1)
	if hasKeywordNow(t, g, id, "trample") {
		t.Error("a misspelled kind is storage, not a keyword counter")
	}
}

// Two counters of one kind are one keyword, and a counter on a creature
// that already has the keyword adds nothing (the Ikoria release notes:
// "multiple instances … are redundant").
func TestSeveralCountersOfOneKindAreOneKeyword(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	addCounterForTest(t, g, id, CounterLifelink, 2)
	addCounterForTest(t, g, id, CounterLifelink, 1)
	registerScopedEffectForTest(t, g, id, []Mod{AddKeywordsMod("lifelink")}, IndefiniteDuration())
	if n := countOf(scopedEffectChar(t, g, id).Abilities, "lifelink"); n != 1 {
		t.Errorf("lifelink appears %d times, want 1", n)
	}
}

// CR 613.3 / 613.7c: a "loses all abilities" OLDER than the counter
// leaves the keyword; one NEWER than it takes the keyword away. The
// counter stays on the permanent either way.
func TestLosesAllAbilitiesIsOrderedAgainstTheCounterByTimestamp(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID

	earlier := pushScopedTestCreature(g, me, 2, 2)
	registerScopedEffectForTest(t, g, earlier, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	addCounterForTest(t, g, earlier, CounterFlying, 1)
	if !hasKeywordNow(t, g, earlier, "flying") {
		t.Error("removal first, counter second: the counter's keyword applies after the removal")
	}

	later := pushScopedTestCreature(g, me, 2, 2)
	addCounterForTest(t, g, later, CounterFlying, 1)
	registerScopedEffectForTest(t, g, later, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	if hasKeywordNow(t, g, later, "flying") {
		t.Error("counter first, removal second: the later removal takes the keyword")
	}
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(later)
		if c.Counters[CounterFlying] != 1 {
			t.Errorf("counters = %v: losing the keyword does not remove the counter", c.Counters)
		}
	})

	// A plain "loses flying" is ordered the same way.
	lose := pushScopedTestCreature(g, me, 2, 2)
	addCounterForTest(t, g, lose, CounterFlying, 1)
	registerScopedEffectForTest(t, g, lose, []Mod{RemoveKeywordsMod("flying")}, IndefiniteDuration())
	if hasKeywordNow(t, g, lose, "flying") {
		t.Error("a later \"loses flying\" takes the counter's flying")
	}
}

// The Ikoria release notes: "Removing a keyword counter doesn't change
// the timestamp of any remaining counters." A placement does
// (CR 613.7c).
func TestRemovingACounterDoesNotRestampAndPlacingOneDoes(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	addCounterForTest(t, g, id, CounterMenace, 2)
	var first int64
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		first = c.CounterStampedAt[CounterMenace]
	})
	if first == 0 {
		t.Fatal("placing a keyword counter must stamp it")
	}
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())

	addCounterForTest(t, g, id, CounterMenace, -1)
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		if got := c.CounterStampedAt[CounterMenace]; got != first {
			t.Errorf("stamp after a removal = %d, want it unchanged at %d", got, first)
		}
	})
	if hasKeywordNow(t, g, id, "menace") {
		t.Error("the removal kept the counter older than the ability removal")
	}

	addCounterForTest(t, g, id, CounterMenace, 1)
	if !hasKeywordNow(t, g, id, "menace") {
		t.Error("a new counter of the kind renews EVERY counter's timestamp, past the removal")
	}

	addCounterForTest(t, g, id, CounterMenace, -2)
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		if _, ok := c.CounterStampedAt[CounterMenace]; ok {
			t.Errorf("stamps = %v: the kind's last counter takes its stamp with it", c.CounterStampedAt)
		}
	})
	if hasKeywordNow(t, g, id, "menace") {
		t.Error("with no counter left there is no keyword")
	}
}

// Proliferate is a placement, so it renews the timestamp (CR 701.34a,
// 613.7c): a keyword a later "loses all abilities" took comes back.
func TestProliferateRenewsAKeywordCountersTimestamp(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := pushScopedTestCreature(g, me, 2, 2)
	addCounterForTest(t, g, id, CounterTrample, 1)
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	if hasKeywordNow(t, g, id, "trample") {
		t.Fatal("setup: the later removal takes trample")
	}
	var err error
	g.WithWriteLock(func() { err = g.ProliferateForEffect(me, uuid.Nil, []uuid.UUID{id}, nil) })
	if err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		if c.Counters[CounterTrample] != 2 {
			t.Fatalf("counters = %v, want two trample counters after proliferate", c.Counters)
		}
	})
	if !hasKeywordNow(t, g, id, "trample") {
		t.Error("proliferated counters carry a new timestamp, later than the removal")
	}
}

// CR 113.11: "It's also impossible for an effect or keyword counter to
// add that ability to the object."
func TestCantHaveBeatsAKeywordCounter(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{CantHaveKeywordsMod("hexproof")}, IndefiniteDuration())
	addCounterForTest(t, g, id, CounterHexproof, 1)
	if hasKeywordNow(t, g, id, "hexproof") {
		t.Error("a can't-have beats a newer keyword counter")
	}
}

// A face-down permanent has no text (CR 708.2a) but keeps its counters,
// and a counter is not text: a face-down 2/2 with a flying counter
// flies.
func TestAFaceDownPermanentKeepsItsCountersKeywords(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	addCounterForTest(t, g, id, CounterFlying, 1)
	var turned []uuid.UUID
	g.WithWriteLock(func() { turned = g.TurnFaceDownForEffect(uuid.Nil, id) })
	if len(turned) != 1 {
		t.Fatalf("TurnFaceDownForEffect turned %v, want the one Bear", turned)
	}
	ch := scopedEffectChar(t, g, id)
	if ch.Name != "" {
		t.Fatalf("setup: the permanent is face down, name %q", ch.Name)
	}
	if !hasAbility(ch.Abilities, "flying") {
		t.Errorf("abilities = %v: the face-down 2/2 keeps its flying counter's keyword", ch.Abilities)
	}
}

// Last-known information (CR 608.2h) is the permanent as it last
// existed, keyword counters included.
func TestLastKnownInformationCarriesTheCountersKeyword(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	addCounterForTest(t, g, id, CounterDeathtouch, 1)
	scopedEffectChar(t, g, id) // settle the layer cache before it leaves
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
	if err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		info, ok := g.LastKnownPermanentForEffect(id)
		if !ok {
			t.Fatal("no last-known record for the destroyed creature")
		}
		if !hasAbility(info.Characteristic.Abilities, "deathtouch") {
			t.Errorf("last-known abilities = %v, want deathtouch", info.Characteristic.Abilities)
		}
	})
}

// CR 122.2: the counters, and so their stamps, do not survive a zone
// change.
func TestAZoneChangeClearsTheStamps(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	addCounterForTest(t, g, id, CounterReach, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatal("the destroyed Bear is nowhere")
		}
		if c.Counters != nil || c.CounterStampedAt != nil {
			t.Errorf("counters %v, stamps %v: both must go with the zone change", c.Counters, c.CounterStampedAt)
		}
	})
}

// CR 122.1b's other half, ADR 0101 Decision 4: a keyword counter on a
// card in another zone gives it the keyword, through HasKeyword and
// through Effective (which the view reads).
func TestAKeywordCounterOffTheBattlefieldGivesItsKeyword(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(Card{
		InstanceID: id, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{CounterShadow: 1, "+1/+1": 1},
	})
	c := me.Graveyard.Cards[len(me.Graveyard.Cards)-1]
	if !HasKeyword(&c, "shadow") {
		t.Error("HasKeyword: a shadow counter on a graveyard card gives it shadow")
	}
	if !hasAbility(c.Effective().Abilities, "shadow") {
		t.Errorf("Effective().Abilities = %v, want shadow", c.Effective().Abilities)
	}
	if got := KeywordCounterTokens(c); len(got) != 1 || got[0] != "shadow" {
		t.Errorf("KeywordCounterTokens = %v, want [shadow]", got)
	}
}

// A keyword counter with no stamp — a restore point written before ADR
// 0101 — grants its keyword, ordered at the permanent's own timestamp
// (owner decision 4): a removal registered after the permanent entered
// is later than it.
func TestAnUnstampedKeywordCounterIsOrderedAtItsPermanent(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	plain := pushScopedTestCreature(g, me, 2, 2)
	removed := pushScopedTestCreature(g, me, 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{plain, removed} {
			i := findCardOnBattlefield(g, id)
			g.Battlefield.Cards[i].Counters = map[string]int{CounterVigilance: 1}
			g.Battlefield.Cards[i].CounterStampedAt = nil
		}
		g.layerVersion.Add(1)
	})
	if !hasKeywordNow(t, g, plain, "vigilance") {
		t.Error("an unstamped keyword counter still grants its keyword")
	}
	registerScopedEffectForTest(t, g, removed, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	if hasKeywordNow(t, g, removed, "vigilance") {
		t.Error("ordered at its permanent's timestamp, the unstamped counter is older than the removal")
	}
}

// Undo is a clone: the stamps ride it, and the clone does not share
// the live card's map.
func TestCloneCopiesTheStamps(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	addCounterForTest(t, g, id, CounterHaste, 1)
	cl := g.Clone()
	var live, copied map[string]int64
	g.WithWriteLock(func() {
		c, _ := g.battlefieldCardLocked(id)
		live = c.CounterStampedAt
	})
	cl.WithWriteLock(func() {
		c, _ := cl.battlefieldCardLocked(id)
		copied = c.CounterStampedAt
	})
	if copied[CounterHaste] == 0 || copied[CounterHaste] != live[CounterHaste] {
		t.Fatalf("clone stamps %v, live %v", copied, live)
	}
	cl.WithWriteLock(func() {
		cl.Battlefield.Cards[findCardOnBattlefield(cl, id)].CounterStampedAt[CounterHaste] = 1
	})
	if live[CounterHaste] == 1 {
		t.Error("the clone shares the live card's stamp map")
	}
}

// The stamps are carried by the snapshot, so a restored table orders a
// keyword counter against a removal exactly as the live one did.
func TestKeywordCounterOrderSurvivesTheSnapshot(t *testing.T) {
	steppedClock(t)
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	addCounterForTest(t, g, id, CounterFirstStrike, 1)
	if !hasKeywordNow(t, g, id, "first strike") {
		t.Fatal("setup: a counter newer than the removal grants first strike")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if !hasKeywordNow(t, restored, id, "first strike") {
		t.Error("restored: the counter is still newer than the removal")
	}
}

func TestKeywordCounterKindsIsSortedAndClosed(t *testing.T) {
	kinds := KeywordCounterKinds()
	if !sort.StringsAreSorted(kinds) {
		t.Errorf("KeywordCounterKinds() = %v, want sorted", kinds)
	}
	kinds[0] = "mutated"
	if IsKeywordCounter("mutated") {
		t.Error("KeywordCounterKinds must return a copy")
	}
	if len(KeywordCounterKinds()) != 15 {
		t.Errorf("%d kinds, want 15 (all of CR 122.1b's fifteen)", len(KeywordCounterKinds()))
	}
}
