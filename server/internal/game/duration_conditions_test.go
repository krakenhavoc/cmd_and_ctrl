package game

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// duration_conditions_test.go is ADR 0109 §2's condition half and §3
// (#1604, #1894): counter-held durations, durations about the pinned
// object, the conjunction, and the restore check on every stored
// duration (Shared machinery 2).

// readyLayers runs the layer pass the way every read does, so a sweep
// at its top sees what the previous mutation changed.
func readyLayers(g *Game) {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
}

func TestACounterHeldDurationEndsWithTheLastCounter(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	land := pushTappedPermanent(g, me, "Plains", "", "Basic Land — Plains", false)
	g.WithWriteLock(func() {
		if _, ok := g.ForAsLongAsPinnedHasCounterDuration(land, "flood"); ok {
			t.Fatal("a land with no flood counter started a counter-held duration (CR 611.2b)")
		}
		_ = g.AddCounterForEffect(land, "flood", 2)
		d, ok := g.ForAsLongAsPinnedHasCounterDuration(land, "flood")
		if !ok {
			t.Fatal("a land with flood counters did not start the duration")
		}
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land), []Mod{AddSubtypesMod("Island")}, d, "Aquitect's Will") {
			t.Fatal("registered nothing")
		}
	})
	if c := scopedEffectChar(t, g, land); !typeListHas(c.Subtypes, "Island") {
		t.Fatalf("with flood counters: subtypes %v, want Island", c.Subtypes)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, "flood", -1) })
	if c := scopedEffectChar(t, g, land); !typeListHas(c.Subtypes, "Island") {
		t.Fatalf("one flood counter left: subtypes %v, want Island", c.Subtypes)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, "flood", -1) })
	if c := scopedEffectChar(t, g, land); typeListHas(c.Subtypes, "Island") {
		t.Fatalf("no flood counter left: subtypes %v, the effect should be over", c.Subtypes)
	}
	// Over for good (CR 611.2b): a new flood counter does not revive it.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, "flood", 1) })
	if c := scopedEffectChar(t, g, land); typeListHas(c.Subtypes, "Island") {
		t.Fatalf("a new flood counter revived the effect: subtypes %v", c.Subtypes)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects left, want 0", n)
	}
}

func TestAConjunctionEndsWhenEitherHalfStops(t *testing.T) {
	cases := map[string]func(g *Game, source uuid.UUID, opp uuid.UUID){
		"the source untaps": func(g *Game, source, _ uuid.UUID) {
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(source) })
		},
		"you lose control of the source": func(g *Game, source, opp uuid.UUID) {
			g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, source, opp, IndefiniteDuration(), "theft") })
		},
	}
	for name, stop := range cases {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0].ID, g.Seats[1].ID
			source := pushTappedPermanent(g, me, "Seasinger", "", "Creature — Merfolk", true)
			victim := pushScopedTestCreature(g, opp, 2, 2)
			g.WithWriteLock(func() {
				d, ok := g.ForAsLongAsYouControlAndSourceTappedDuration(source, me)
				if !ok {
					t.Fatal("a tapped source you control did not start the duration")
				}
				if len(d.Also) != 1 || d.Also[0] != WhileSourceRemainsTapped || d.Condition != WhileYouControlSource {
					t.Fatalf("duration %+v, want WhileYouControlSource and WhileSourceRemainsTapped", d)
				}
				g.GainControlForEffect(source, victim, me, d, "Seasinger")
			})
			readyLayers(g)
			if got := controllerOfCard(t, g, victim); got != me {
				t.Fatalf("setup: controller %s, want %s", got, me)
			}
			stop(g, source, opp)
			readyLayers(g)
			if got := controllerOfCard(t, g, victim); got != opp {
				t.Errorf("%s and the creature stayed stolen (CR 611.2b)", name)
			}
		})
	}
}

func TestAConjunctionNeverStartsWhenEitherHalfIsFalse(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	upright := pushTappedPermanent(g, me, "Upright", "", "Creature", false)
	theirs := pushTappedPermanent(g, opp, "Theirs", "", "Creature", true)
	g.WithWriteLock(func() {
		if _, ok := g.ForAsLongAsYouControlAndSourceTappedDuration(upright, me); ok {
			t.Error("an untapped source started the duration")
		}
		if _, ok := g.ForAsLongAsYouControlAndSourceTappedDuration(theirs, me); ok {
			t.Error("a source you don't control started the duration")
		}
	})
}

func TestAPinnedTappedDurationEndsWhenThePinUntaps(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	zygon := pushScopedTestCreature(g, me, 3, 2)
	model := pushTappedPermanent(g, opp, "Tapped", "", "Creature — Bear", true)
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsPinnedTappedDuration(model)
		if !ok {
			t.Fatal("a tapped permanent did not start the duration")
		}
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(zygon), []Mod{ModifyPTMod(1, 1)}, d, "Zygon")
	})
	if c := scopedEffectChar(t, g, zygon); c.Power != 4 {
		t.Fatalf("while tapped: power %d, want 4", c.Power)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(model) })
	if c := scopedEffectChar(t, g, zygon); c.Power != 3 {
		t.Errorf("after the pin untapped: power %d, want 3", c.Power)
	}
}

func TestAPowerConditionEndsAfterThePassThatBreaksIt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	oldMan := pushScopedTestCreature(g, me, 2, 3)
	g.WithWriteLock(func() { cardByIDForUntapTest(g, oldMan).Tapped = true })
	victim := pushScopedTestCreature(g, opp, 2, 2)
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsSourceTappedAndPowerAtMostDuration(oldMan, victim)
		if !ok {
			t.Fatal("a 2-power creature under a tapped 2-power source did not start the duration")
		}
		g.GainControlForEffect(oldMan, victim, me, d, "Old Man of the Sea")
	})
	readyLayers(g)
	if got := controllerOfCard(t, g, victim); got != me {
		t.Fatalf("setup: controller %s, want %s", got, me)
	}
	// Giant Growth on the stolen creature: its power is a layer output,
	// so the pass that raises it is followed by one that ends the theft.
	registerScopedEffectForTest(t, g, victim, []Mod{ModifyPTMod(3, 3)}, IndefiniteDuration())
	readyLayers(g)
	readyLayers(g)
	if got := controllerOfCard(t, g, victim); got != opp {
		t.Errorf("the stolen creature's power passed the source's and it stayed stolen")
	}
	// And a 3-power creature never starts under a 2-power source.
	big := pushScopedTestCreature(g, opp, 3, 3)
	g.WithWriteLock(func() {
		if _, ok := g.ForAsLongAsSourceTappedAndPowerAtMostDuration(oldMan, big); ok {
			t.Error("a 3-power creature started the duration under a 2-power source")
		}
	})
}

// TestANewDurationRoundTrips: the new fields are data, so a table
// holding one is a restore point and comes back with them.
func TestANewDurationRoundTrips(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	source := pushTappedPermanent(g, me, "Seasinger", "", "Creature — Merfolk", true)
	victim := pushScopedTestCreature(g, opp, 2, 2)
	land := pushTappedPermanent(g, opp, "Plains", "", "Basic Land — Plains", false)
	g.WithWriteLock(func() {
		d, _ := g.ForAsLongAsYouControlAndSourceTappedDuration(source, me)
		g.GainControlForEffect(source, victim, me, d, "Seasinger")
		_ = g.AddCounterForEffect(land, "flood", 1)
		fd, _ := g.ForAsLongAsPinnedHasCounterDuration(land, "flood")
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land), []Mod{AddSubtypesMod("Island")}, fd, "flood")
	})
	readyLayers(g)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if n := len(restored.ScopedEffects); n != 2 {
		t.Fatalf("%d scoped effects restored, want 2", n)
	}
	for i, e := range g.ScopedEffects {
		if !restored.ScopedEffects[i].Duration.Equal(e.Duration) {
			t.Errorf("record %d: duration %+v restored as %+v", i, e.Duration, restored.ScopedEffects[i].Duration)
		}
	}
	restored.WithWriteLock(func() { _ = restored.UntapTargetForEffect(source) })
	readyLayers(restored)
	if got := controllerOfCard(t, restored, victim); got != opp {
		t.Error("the restored conjunction did not end when the source untapped")
	}
}

// TestDurationRoutesCoverEveryStoredDuration holds the restore check to
// the snapshot's real shape: every path in the recorded shape at which
// a Duration sits must be one the check walks. A new field that stores
// a duration is covered automatically (the routes are reflected), and
// this test says so in a form an operator can read.
func TestDurationRoutesCoverEveryStoredDuration(t *testing.T) {
	routes := map[string]bool{}
	for _, r := range durationRoutes() {
		routes[r] = true
	}
	for _, want := range []string{
		"scopedEffects[].duration",
		"delayedTriggers[].duration",
		"seats[].castPermissions[].duration",
		"seats[].statics[].duration",
		"battlefield.cards[].nextUntapSkips[].while",
		"exile.cards[].nextUntapSkips[].while",
		"phasedOut.cards[].nextUntapSkips[].while",
		"stack.cards[].nextUntapSkips[].while",
		"lastKnownStack[].card.nextUntapSkips[].while",
		"seats[].hand.cards[].nextUntapSkips[].while",
		"seats[].library.cards[].nextUntapSkips[].while",
		"seats[].graveyard.cards[].nextUntapSkips[].while",
		"seats[].command.cards[].nextUntapSkips[].while",
		"seats[].emblems.cards[].nextUntapSkips[].while",
	} {
		if !routes[want] {
			t.Errorf("the restore check does not walk %s", want)
		}
	}
	// Every object in the recorded shape that has a Duration's
	// signature keys is a route.
	f, err := os.Open(shapeGoldenPath(SnapshotSchemaVersion))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	kinds, turns := map[string]bool{}, map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		path, _, _ := strings.Cut(sc.Text(), "\t")
		if p, ok := strings.CutSuffix(path, ".ExpiresAtTurnsBegun"); ok {
			turns[p] = true
		}
		if p, ok := strings.CutSuffix(path, ".Kind"); ok {
			kinds[p] = true
		}
	}
	for p := range turns {
		if kinds[p] && !routes[p] {
			t.Errorf("the shape records a duration at %s that the restore check does not walk", p)
		}
	}
}

// TestAnUnreadableDurationIsRefusedWherever is ADR 0109 Shared
// machinery 2: before it, only scopedEffects and delayedTriggers were
// checked, and an older binary read an unknown condition on a cast
// permission, a player static or an untap hold as "while the source is
// on the battlefield". Every stored duration is now refused, by both
// restores, when this binary cannot read it.
func TestAnUnreadableDurationIsRefusedWherever(t *testing.T) {
	future := Duration{Kind: ForAsLongAs, Condition: DurationCondition(42)}
	cases := map[string]func(s *GameSnapshot){
		"cast permission": func(s *GameSnapshot) {
			s.Seats[0].CastPermissions = append(s.Seats[0].CastPermissions, CastPermission{Duration: future})
		},
		"player static": func(s *GameSnapshot) {
			s.Seats[0].Statics = append(s.Seats[0].Statics, PlayerStatic{Duration: future})
		},
		"untap hold on the battlefield": func(s *GameSnapshot) {
			d := future
			s.Battlefield.Cards[0].NextUntapSkips = []untapSkipSnapshot{{While: &d}}
		},
		"delayed trigger": func(s *GameSnapshot) {
			d := future
			s.DelayedTriggers[0].Duration = &d
		},
		"scoped effect, a joined condition": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration = Duration{Kind: ForAsLongAs, Also: []DurationCondition{42}}
		},
		"Also on an until-end-of-turn duration": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration = Duration{Kind: UntilEndOfTurn, Also: []DurationCondition{WhileSourceRemainsTapped}}
		},
		"a counter-held duration with no counter kind": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration = Duration{Kind: ForAsLongAs, Condition: WhilePinnedHasCounter, Pinned: uuid.New()}
		},
		"a counter kind with no counter condition": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration = Duration{Kind: ForAsLongAs, CounterKind: "flood"}
		},
		"a pinned condition with no pin": func(s *GameSnapshot) {
			s.ScopedEffects[0].Duration = Duration{Kind: ForAsLongAs, Condition: WhilePinnedRemainsTapped}
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			loose, strict := corruptAndRestore(t, corrupt)
			for which, err := range map[string]error{"Restore": loose, "RestoreStrict": strict} {
				if !errors.Is(err, ErrUnknownEffectKey) {
					t.Errorf("%s: err = %v, want ErrUnknownEffectKey", which, err)
				}
			}
		})
	}
}

// TestDurationPhrasesNameEveryCondition: the land-type chip reads a
// duration in the card's words, every condition of a conjunction
// included.
func TestDurationPhrasesNameEveryCondition(t *testing.T) {
	g := newActiveGame(t)
	cases := map[string]Duration{
		"for as long as you control Seasinger and Seasinger remains tapped": {
			Kind: ForAsLongAs, Condition: WhileYouControlSource, Also: []DurationCondition{WhileSourceRemainsTapped}},
		"for as long as it has a flood counter on it": {
			Kind: ForAsLongAs, Condition: WhilePinnedHasCounter, CounterKind: "flood"},
		"for as long as it has an awakening counter on it": {
			Kind: ForAsLongAs, Condition: WhilePinnedHasCounter, CounterKind: "awakening"},
	}
	for want, d := range cases {
		var got string
		g.WithWriteLock(func() { got = g.durationPhraseLocked(d, "Seasinger") })
		if got != want {
			t.Errorf("phrase %q, want %q", got, want)
		}
	}
}

// TestAnUnknownDurationFieldIsRefusedWherever is the field half: a key
// a newer Duration carries is refused on every stored duration, not
// only on a scoped effect's. encoding/json would drop it, and the
// effect would outlast the condition the key named.
func TestAnUnknownDurationFieldIsRefusedWherever(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	source := pushTappedPermanent(g, me, "Source", "", "Creature", false)
	target := pushTappedPermanent(g, opp, "Held", "", "Creature", true)
	holdWhileYouControl(t, g, source, me, target)
	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(me, "hexproof", "test static", source, IndefiniteDuration())
	})
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	future := func(d map[string]any) { d["Unless"] = 3 }
	cases := map[string]func(snap map[string]any){
		"untap hold": func(snap map[string]any) {
			for _, c := range snap["battlefield"].(map[string]any)["cards"].([]any) {
				card := c.(map[string]any)
				if skips, ok := card["nextUntapSkips"].([]any); ok {
					future(skips[0].(map[string]any)["while"].(map[string]any))
				}
			}
		},
		"player static": func(snap map[string]any) {
			seat := snap["seats"].([]any)[0].(map[string]any)
			future(seat["statics"].([]any)[0].(map[string]any)["duration"].(map[string]any))
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			var generic map[string]any
			if err := json.Unmarshal(raw, &generic); err != nil {
				t.Fatal(err)
			}
			corrupt(generic)
			bad, err := json.Marshal(generic)
			if err != nil {
				t.Fatal(err)
			}
			var snap GameSnapshot
			if err := json.Unmarshal(bad, &snap); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("Restore: err = %v, want ErrUnknownEffectKey", err)
			}
			if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
				t.Errorf("RestoreStrict: err = %v, want ErrUnknownEffectKey", err)
			}
		})
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if _, err := snap.RestoreStrict(); err != nil {
		t.Errorf("an intact file was refused: %v", err)
	}
}
