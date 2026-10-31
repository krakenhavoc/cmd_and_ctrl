package game

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mana_trigger_judge_test.go — ADR 0074's amendment of 2026-10-30
// (#4716), slice 1: pass one of fireManaTriggersLocked is the read-only
// judge manaTriggerFiringsLocked, with manaTriggersExistLocked as the
// cheap gate in front of it. The executor's behaviour is pinned by
// mana_trigger_test.go and mana_trigger_controller_test.go, which pass
// unchanged; these pin the new seam.

const (
	judgeSilentOracle    = "test-judge-silent"
	judgeSwampOnlyOracle = "test-judge-swamp-only"
	judgeSkipperOracle   = "test-judge-skipper"
	judgeYoursOracle     = "test-judge-yours"
)

// judgeTriggers is a catalog with one of each shape the judge has to
// tell apart: an Aura that fires, a global trigger that reads the
// produced colour, one that adds nothing, one the production does not
// satisfy, one with Winter's Night's rider, and one that only sees its
// own controller's taps.
func judgeTriggers(key string) []ManaTrigger {
	landTapped := func(prod ManaProduced, _ *Card, _ *Game) bool { return prod.Source.IsLand() }
	switch key {
	case wildGrowthTestOracle:
		return []ManaTrigger{wildGrowthTrigger("{G}")}
	case manaFlareTestOracle:
		return []ManaTrigger{{
			Label:     "Test Flare — one mana of any type that land produced",
			AppliesTo: landTapped,
			Produced:  func(prod ManaProduced, _ *Card, _ *Game) string { return prod.ProducedColorPipe() },
		}}
	case judgeSilentOracle:
		return []ManaTrigger{{
			Label:     "Test Silent — adds nothing",
			AppliesTo: landTapped,
			Produced:  func(ManaProduced, *Card, *Game) string { return "" },
		}}
	case judgeSwampOnlyOracle:
		return []ManaTrigger{{
			Label:     "Test Swamp — add {B} for a Swamp",
			AppliesTo: func(prod ManaProduced, _ *Card, _ *Game) bool { return prod.Source.HasSubtype("Swamp") },
			Produced:  func(ManaProduced, *Card, *Game) string { return "{B}" },
		}}
	case judgeSkipperOracle:
		return []ManaTrigger{{
			Label:                "Test Skipper — add {U}, the land stays tapped",
			AppliesTo:            landTapped,
			Produced:             func(ManaProduced, *Card, *Game) string { return "{U}" },
			TappedSkipsNextUntap: true,
		}}
	case judgeYoursOracle:
		return []ManaTrigger{{
			Label: "Test Yours — add {R} when you tap a land",
			AppliesTo: func(prod ManaProduced, source *Card, _ *Game) bool {
				return prod.Source.IsLand() && prod.Controller == source.Controller
			},
			Produced: func(ManaProduced, *Card, *Game) string { return "{R}" },
		}}
	}
	return nil
}

// judgeBoard is the board the judge tests share. The Forest's controller
// is `me`; the opponent owns the Mana Flare, the Skipper and a "you"
// trigger, and three delayed records say what a spell set up this turn.
type judgeBoard struct {
	g          *Game
	me, opp    *Player
	forest     uuid.UUID
	growth     uuid.UUID
	flare      uuid.UUID
	skipper    uuid.UUID
	delayedSrc uuid.UUID
}

func newJudgeBoard(t *testing.T) *judgeBoard {
	t.Helper()
	g := newActiveGame(t)
	withCatalogManaTriggers(t, judgeTriggers)
	b := &judgeBoard{g: g, me: g.Seats[0], opp: g.Seats[1]}
	b.forest = pushForest(g, b.me)
	b.growth = pushAuraOn(g, b.me, "Wild Growth", wildGrowthTestOracle, b.forest)
	b.flare = pushBattlefieldForTest(g, b.opp.ID, "Mana Flare", "Enchantment", manaFlareTestOracle)
	pushBattlefieldForTest(g, b.me.ID, "Silent", "Enchantment", judgeSilentOracle)
	pushBattlefieldForTest(g, b.me.ID, "Swamp Only", "Enchantment", judgeSwampOnlyOracle)
	b.skipper = pushBattlefieldForTest(g, b.me.ID, "Skipper", "Enchantment", judgeSkipperOracle)
	pushBattlefieldForTest(g, b.opp.ID, "Yours", "Enchantment", judgeYoursOracle)

	b.delayedSrc = uuid.New()
	g.WithWriteLock(func() {
		// Fires: any player's Forest this turn.
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: b.opp.ID, SourceCardID: b.delayedSrc, Label: "Test High Tide",
			ManaTapSubtype: "Forest", ManaAdds: "{R}", Repeats: true,
		})
		// Does not fire: "you" is the opponent, and the Forest is mine.
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: b.opp.ID, SourceCardID: uuid.New(), Label: "Test Molten Tide",
			ManaTapSubtype: "Forest", ManaAdds: "{W}", Repeats: true, ManaTapControllerOnly: true,
		})
		// Does not fire: the wrong land type.
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: b.me.ID, SourceCardID: uuid.New(), Label: "Test Bubbling Muck",
			ManaTapSubtype: "Swamp", ManaAdds: "{B}", Repeats: true,
		})
	})
	return b
}

// production is my Forest tapped for {G}, as the executor reports it.
func (b *judgeBoard) production(t *testing.T) ManaProduced {
	t.Helper()
	c := b.g.findBattlefieldCardLocked(b.forest)
	if c == nil {
		t.Fatal("the Forest is not on the battlefield")
	}
	return ManaProduced{Source: *c, Controller: b.me.ID, Colors: []string{"G"}}
}

// judge asks the judge under a read lock, the way the planner will.
func (b *judgeBoard) judge(prod ManaProduced) []manaTriggerFiring {
	var out []manaTriggerFiring
	b.g.ReadSnapshot(func() { out = b.g.manaTriggerFiringsLocked(prod) })
	return out
}

// persisted is the game's persistence format as bytes, with the one
// field that is the wall clock zeroed.
func persisted(t *testing.T, g *Game) []byte {
	t.Helper()
	s := g.CaptureSnapshot()
	s.TakenAt = time.Time{}
	blob, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	return blob
}

// The judge's list is what the executor resolves: same sources, same
// order, same mana. The delayed records are part of it.
func TestManaTriggerFiringsMatchWhatFires(t *testing.T) {
	b := newJudgeBoard(t)
	prod := b.production(t)

	judged := b.judge(prod)
	wantSources := []uuid.UUID{b.delayedSrc, b.growth, b.flare, b.skipper}
	if len(judged) != len(wantSources) {
		t.Fatalf("judge found %d firings %+v, want %d: the delayed High Tide, Wild Growth, the Flare and the Skipper",
			len(judged), judged, len(wantSources))
	}
	for i, want := range wantSources {
		if judged[i].source != want {
			t.Errorf("firing %d is from %s (%s), want %s: delayed records first, then battlefield order",
				i, judged[i].source, judged[i].label, want)
		}
		if judged[i].produced == "" {
			t.Errorf("firing %d (%s) adds nothing: a silent ability is not a firing", i, judged[i].label)
		}
		if got, want := judged[i].skip, judged[i].source == b.skipper; got != want {
			t.Errorf("firing %d (%s) skip = %v, want %v", i, judged[i].label, got, want)
		}
	}

	type token struct {
		source uuid.UUID
		color  string
	}
	want := map[token]int{}
	for _, f := range judged {
		slots, err := ParseProducedMana(f.produced)
		if err != nil {
			t.Fatalf("judge returned %q, which does not parse: %v", f.produced, err)
		}
		for _, s := range slots {
			if len(s.Options) != 1 {
				t.Fatalf("this fixture only uses fixed colours, got %+v", s)
			}
			want[token{f.source, s.Options[0]}]++
		}
	}

	b.g.WithWriteLock(func() { b.g.fireManaTriggersLocked(prod, nil, nil) })
	got := map[token]int{}
	for _, tok := range b.me.ManaPool.Units() {
		got[token{tok.Source, tok.Color}]++
	}
	if len(got) != len(want) {
		t.Errorf("the executor minted %v, the judge predicted %v", got, want)
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("the executor minted %d of %v, the judge predicted %d (pool %v)", got[k], k, n, got)
		}
	}
	// The rider is the executor's to apply, not the judge's.
	if c := b.g.findBattlefieldCardLocked(b.forest); c == nil || len(c.NextUntapSkips) != 1 {
		t.Errorf("the Forest should skip its next untap once, got %+v", c)
	}
}

// The judge changes nothing: not the pool, not the delayed queue, not a
// permanent, not the event log. Asking it twice gives the same answer.
func TestManaTriggerJudgeChangesNothing(t *testing.T) {
	b := newJudgeBoard(t)
	prod := b.production(t)
	b.g.ReadSnapshot(func() {})

	before := persisted(t, b.g)
	first := b.judge(prod)
	second := b.judge(prod)
	after := persisted(t, b.g)

	if string(before) != string(after) {
		t.Error("the judge changed the game's snapshot")
	}
	if len(first) == 0 {
		t.Fatal("the fixture should fire something, or this proves nothing")
	}
	if len(first) != len(second) {
		t.Fatalf("two asks gave %d and %d firings", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("firing %d differs between asks: %+v vs %+v", i, first[i], second[i])
		}
	}
	if len(b.me.ManaPool.Units()) != 0 {
		t.Errorf("the judge put mana in the pool: %v", b.me.ManaPool.Units())
	}
	if c := b.g.findBattlefieldCardLocked(b.forest); c == nil || len(c.NextUntapSkips) != 0 {
		t.Errorf("the judge applied the Skipper's rider: %+v", c)
	}
}

// A production that adds nothing, or lands in nobody's pool, fires
// nothing, however many triggers are out.
func TestManaTriggerJudgeIgnoresAProductionThatAddsNothing(t *testing.T) {
	b := newJudgeBoard(t)
	prod := b.production(t)

	empty := prod
	empty.Colors = nil
	if got := b.judge(empty); len(got) != 0 {
		t.Errorf("a production of nothing fired %+v", got)
	}
	nobody := prod
	nobody.Controller = uuid.Nil
	if got := b.judge(nobody); len(got) != 0 {
		t.Errorf("a production into nobody's pool fired %+v", got)
	}
}

// The judge judges the production it is handed, and only that one: the
// planner will ask about a land that is not tapped yet, and about every
// player's.
func TestManaTriggerJudgeJudgesTheProductionItIsGiven(t *testing.T) {
	b := newJudgeBoard(t)
	b.g.WithWriteLock(func() {
		for i := range b.g.Battlefield.Cards {
			if b.g.Battlefield.Cards[i].InstanceID == b.forest {
				b.g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	prod := b.production(t)
	if len(b.judge(prod)) == 0 {
		t.Error("the judge refused a production because its land is already tapped: it judges the production, not the land")
	}

	// An opponent's Swamp, into the opponent's pool: the Flare, the
	// Skipper, Swamp Only, the opponent's own "you" trigger and the
	// delayed Muck (whose "a player" is anyone) see it. My Wild Growth
	// is on my Forest, so it does not.
	swamp := pushBattlefieldForTest(b.g, b.opp.ID, "Swamp", "Basic Land — Swamp", "test-swamp")
	c := b.g.findBattlefieldCardLocked(swamp)
	theirs := ManaProduced{Source: *c, Controller: b.opp.ID, Colors: []string{"B"}}
	for _, f := range b.judge(theirs) {
		if f.source == b.growth {
			t.Errorf("Wild Growth fired for a Swamp it is not on: %+v", f)
		}
	}
	if got := len(b.judge(theirs)); got != 5 {
		t.Errorf("an opponent's Swamp fired %d abilities, want 5: Flare, Skipper, Swamp Only, their own 'you' trigger and the delayed Muck", got)
	}
}

// The gate: false only when nothing could ever fire.
func TestManaTriggersExistLocked(t *testing.T) {
	t.Run("no catalog hook", func(t *testing.T) {
		g := newActiveGame(t)
		pushForest(g, g.Seats[0])
		prev := CatalogManaTriggers
		CatalogManaTriggers = nil
		t.Cleanup(func() { CatalogManaTriggers = prev })
		if g.manaTriggersExistLocked() {
			t.Error("a game with no mana trigger catalog has no mana triggers")
		}
	})
	t.Run("plain board", func(t *testing.T) {
		g := newActiveGame(t)
		withCatalogManaTriggers(t, judgeTriggers)
		pushForest(g, g.Seats[0])
		pushBattlefieldForTest(g, g.Seats[1].ID, "Bear", "Creature — Bear", "test-bear")
		if g.manaTriggersExistLocked() {
			t.Error("a board with no permanent that has a mana trigger has none")
		}
	})
	t.Run("a permanent with one", func(t *testing.T) {
		g := newActiveGame(t)
		withCatalogManaTriggers(t, judgeTriggers)
		pushForest(g, g.Seats[0])
		pushBattlefieldForTest(g, g.Seats[1].ID, "Mana Flare", "Enchantment", manaFlareTestOracle)
		if !g.manaTriggersExistLocked() {
			t.Error("an opponent's Mana Flare is a mana trigger on the board")
		}
	})
	t.Run("a delayed mana record", func(t *testing.T) {
		g := newActiveGame(t)
		withCatalogManaTriggers(t, judgeTriggers)
		pushForest(g, g.Seats[0])
		g.WithWriteLock(func() {
			g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
				Controller: g.Seats[0].ID, SourceCardID: uuid.New(), Label: "Test Muck",
				ManaTapSubtype: "Swamp", ManaAdds: "{B}", Repeats: true,
			})
		})
		if !g.manaTriggersExistLocked() {
			t.Error("a spell's delayed mana trigger is a mana trigger")
		}
	})
	t.Run("a delayed record that is not a mana trigger", func(t *testing.T) {
		g := newActiveGame(t)
		withCatalogManaTriggers(t, judgeTriggers)
		g.DelayedTriggers = append(g.DelayedTriggers, &DelayedTrigger{Controller: g.Seats[0].ID, Label: "stack trigger"}, nil)
		if g.manaTriggersExistLocked() {
			t.Error("a delayed stack trigger, or a nil record, is not a mana trigger")
		}
	})
	t.Run("abilities removed", func(t *testing.T) {
		g := newActiveGame(t)
		withCatalogManaTriggers(t, judgeTriggers)
		forest := pushForest(g, g.Seats[0])
		aura := pushAuraOn(g, g.Seats[0], "Wild Growth", wildGrowthTestOracle, forest)
		if !g.manaTriggersExistLocked() {
			t.Fatal("Wild Growth is a mana trigger")
		}
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if c := &g.Battlefield.Cards[i]; c.InstanceID == aura {
					eff := c.printedCharacteristic()
					eff.AbilitiesRemoved = true
					c.effective = &eff
				}
			}
		})
		if g.manaTriggersExistLocked() {
			t.Error("an Aura that lost its abilities (CR 613.1f) has no mana trigger")
		}
	})
	t.Run("a closed designation gate", func(t *testing.T) {
		g := newActiveGame(t)
		withCatalogManaTriggers(t, func(key string) []ManaTrigger {
			if key != wildGrowthTestOracle {
				return nil
			}
			gated := wildGrowthTrigger("{G}")
			gated.ActiveWhen = ClassLevel(2)
			return []ManaTrigger{gated}
		})
		forest := pushForest(g, g.Seats[0])
		aura := pushAuraOn(g, g.Seats[0], "Gated Growth", wildGrowthTestOracle, forest)
		if g.manaTriggersExistLocked() {
			t.Error("a level-2 trigger on a level-1 Class has not been unlocked")
		}
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == aura {
					g.Battlefield.Cards[i].ClassLevel = 2
				}
			}
		})
		if !g.manaTriggersExistLocked() {
			t.Error("at level 2 the gate opens")
		}
	})
}

// The gate is sound: wherever it says no, the judge says nothing for
// every production, so a caller may skip the judge on a no.
func TestManaTriggersExistLockedIsSoundForTheJudge(t *testing.T) {
	g := newActiveGame(t)
	withCatalogManaTriggers(t, judgeTriggers)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushBattlefieldForTest(g, g.Seats[1].ID, "Bear", "Creature — Bear", "test-bear")
	prod := ManaProduced{Source: *g.findBattlefieldCardLocked(forest), Controller: me.ID, Colors: []string{"G"}}

	var exists bool
	var firings []manaTriggerFiring
	g.ReadSnapshot(func() {
		exists = g.manaTriggersExistLocked()
		firings = g.manaTriggerFiringsLocked(prod)
	})
	if exists || len(firings) != 0 {
		t.Fatalf("a board with no mana trigger: gate = %v, judge = %+v", exists, firings)
	}

	// The executor takes the gate's no as an answer: it mints nothing.
	g.WithWriteLock(func() { g.fireManaTriggersLocked(prod, nil, nil) })
	if got := me.ManaPool.Units(); len(got) != 0 {
		t.Errorf("the executor minted %v on a board with no trigger", got)
	}
}
