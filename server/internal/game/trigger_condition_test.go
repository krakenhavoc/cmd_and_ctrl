package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_condition_test.go — #3033: an intervening "if" as a footprint
// step (FootprintCondition). Each condition kind is read as
// addConditionReadLocked argues, and conflicts only with what can change
// its answer.

// condFixture is a two-seat game with one creature of mine on the
// battlefield, the source of every item.
type condFixture struct {
	g       *Game
	me, opp *Player
	src     uuid.UUID
}

func newCondFixture(t *testing.T) condFixture {
	t.Helper()
	g := newActiveGame(t)
	f := condFixture{g: g, me: g.Seats[0], opp: g.Seats[1], src: uuid.New()}
	g.Battlefield.PushTop(Card{
		InstanceID: f.src, Name: "Condition Source", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: f.me.ID, Controller: f.me.ID,
	})
	return f
}

func (f condFixture) item() *StackItem {
	return &StackItem{Kind: StackItemTriggered, ID: uuid.New(), Controller: f.me.ID, SourceCardID: f.src}
}

// access resolves steps against a fresh item of mine.
func (f condFixture) access(t *testing.T, steps ...FootprintStep) (triggerAccess, bool) {
	t.Helper()
	item := f.item()
	a := newTriggerAccess(item.Controller)
	a.source = item.SourceCardID
	for _, s := range steps {
		if !f.g.addFootprintStepLocked(&a, item, s, nil) {
			return triggerAccess{}, false
		}
	}
	f.g.markRulesLocked(&a)
	return a, true
}

func (f condFixture) mustAccess(t *testing.T, steps ...FootprintStep) triggerAccess {
	t.Helper()
	a, ok := f.access(t, steps...)
	if !ok {
		t.Fatalf("the check refused %+v", steps)
	}
	return a
}

// conflict is the reason a and b conflict, either way round.
func (f condFixture) conflict(a, b triggerAccess) TriggerOrderReason {
	statics := func() layerStatics { return f.g.layerStaticsLocked() }
	if w := f.g.accessesConflictLocked(&a, &b, map[uuid.UUID]bool{}, statics); w.reason != "" {
		return w.reason
	}
	return f.g.accessesConflictLocked(&b, &a, map[uuid.UUID]bool{}, statics).reason
}

func cond(c FootprintCond, n int) FootprintStep {
	return FootprintStep{Kind: FootprintCondition, Cond: c, N: n}
}

var (
	drawOne = FootprintStep{Kind: FootprintDraw, N: 1}
	gainOne = FootprintStep{Kind: FootprintGainLife, N: 1}
)

// A tally that only rises, and the city's blessing, read nothing once
// they hold, and are refused while they don't: another item could make
// them true.
func TestAMonotoneConditionReadsNothingOnceItHolds(t *testing.T) {
	for _, c := range []struct {
		name string
		step FootprintStep
		make func(f condFixture)
	}{
		{"life gained", cond(FootprintCondLifeGained, 3), func(f condFixture) {
			f.g.TurnTally.Players = map[uuid.UUID]PlayerTurnTally{f.me.ID: {LifeGained: 3}}
		}},
		{"life lost", cond(FootprintCondLifeLost, 1), func(f condFixture) {
			f.g.TurnTally.Players = map[uuid.UUID]PlayerTurnTally{f.me.ID: {LifeLost: 1}}
		}},
		{"a creature died", cond(FootprintCondCreatureDied, 0), func(f condFixture) {
			f.g.TurnTally.Players = map[uuid.UUID]PlayerTurnTally{f.me.ID: {CreaturesDied: 1}}
		}},
		{"the city's blessing", cond(FootprintCondCitysBlessing, 0), func(f condFixture) { f.me.CitysBlessing = true }},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newCondFixture(t)
			f.g.TurnTally.Players = nil
			if _, ok := f.access(t, c.step, drawOne); ok {
				t.Fatal("a condition that does not hold yet must be refused")
			}
			c.make(f)
			a := f.mustAccess(t, c.step, drawOne)
			if len(a.reads) != 0 || a.boardReads || a.monarchRead {
				t.Errorf("a held condition read %v (board %v, monarch %v), want nothing", a.reads, a.boardReads, a.monarchRead)
			}
			// It conflicts with nothing the guarded draw would not:
			// another item that gains life and makes a token.
			b := f.mustAccess(t, gainOne, FootprintStep{Kind: FootprintCreateToken, N: 1})
			if r := f.conflict(a, b); r != "" {
				t.Errorf("conflicts under %q", r)
			}
		})
	}
}

// "You cast no spells this turn" can only turn false, by a cast, and no
// step casts.
func TestNoSpellsCastReadsNothingWhileItHolds(t *testing.T) {
	f := newCondFixture(t)
	f.g.SpellsCastThisTurn = nil
	a := f.mustAccess(t, cond(FootprintCondNoSpellsCast, 0), FootprintStep{Kind: FootprintCounterOnSource, N: 2, Counter: CounterPlusOne})
	if a.boardReads || a.monarchRead {
		t.Errorf("read the board or the monarch: %+v", a)
	}
	f.g.SpellsCastThisTurn = map[uuid.UUID]CastTally{f.me.ID: {Total: 1}}
	if _, ok := f.access(t, cond(FootprintCondNoSpellsCast, 0), drawOne); ok {
		t.Error("once a spell was cast the condition is false, and the step is refused")
	}
}

func TestTheMonarchConditionConflictsOnlyWithAMonarchStep(t *testing.T) {
	f := newCondFixture(t)
	a := f.mustAccess(t, cond(FootprintCondMonarch, 0), drawOne)
	if !a.monarchRead {
		t.Fatal("the monarch condition must be a read")
	}
	if r := f.conflict(a, f.mustAccess(t, gainOne)); r != "" {
		t.Errorf("a gain conflicts under %q", r)
	}
	if r := f.conflict(a, f.mustAccess(t, FootprintStep{Kind: FootprintMonarch})); r != TriggerOrderReasonCondition {
		t.Errorf("becoming the monarch: %q, want %q", r, TriggerOrderReasonCondition)
	}
}

// Runaway Steam-Kin's "fewer than three counters on it" reads its
// source, so another item that writes the source ties with it.
func TestTheSourceConditionReadsTheSource(t *testing.T) {
	f := newCondFixture(t)
	a := f.mustAccess(t, cond(FootprintCondSource, 3), drawOne)
	if !a.reads[f.src] {
		t.Fatal("the source condition must read the source")
	}
	b := f.mustAccess(t, FootprintStep{Kind: FootprintCounterOnSource, N: 1, Counter: CounterPlusOne})
	if r := f.conflict(a, b); r != TriggerOrderReasonObjectTie {
		t.Errorf("a counter on the same source: %q, want %q", r, TriggerOrderReasonObjectTie)
	}
}

func TestABoardConditionConflictsWithWhatChangesTheBoard(t *testing.T) {
	f := newCondFixture(t)
	// The guarded step is a gain, so a scry or a mill shares no library
	// with it.
	a := f.mustAccess(t, cond(FootprintCondBoard, 4), gainOne)
	if !a.boardReads {
		t.Fatal("a board condition must read the board")
	}
	for _, c := range []struct {
		name string
		b    FootprintStep
		want TriggerOrderReason
	}{
		{"a draw", drawOne, ""},
		{"a gain", gainOne, ""},
		{"a drain", FootprintStep{Kind: FootprintLoseLifeEachOpponent, N: 1}, ""},
		{"a scry", FootprintStep{Kind: FootprintScry, N: 1}, ""},
		{"a counter on a permanent", FootprintStep{Kind: FootprintCounterOnSource, N: 1, Counter: CounterPlusOne}, TriggerOrderReasonCondition},
		{"a token", FootprintStep{Kind: FootprintCreateToken, N: 1}, TriggerOrderReasonCondition},
		{"a mill", FootprintStep{Kind: FootprintMill, N: 1}, TriggerOrderReasonCondition},
		{"the monarch", FootprintStep{Kind: FootprintMonarch}, TriggerOrderReasonCondition},
		{"energy", FootprintStep{Kind: FootprintEnergy, N: 1}, TriggerOrderReasonCondition},
	} {
		t.Run(c.name, func(t *testing.T) {
			// The other item has its own source, so a counter on it is
			// not a tie with a's source.
			other := uuid.New()
			f.g.Battlefield.PushTop(Card{InstanceID: other, Name: "Other", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Owner: f.me.ID, Controller: f.me.ID})
			item := f.item()
			item.SourceCardID = other
			b := newTriggerAccess(item.Controller)
			if !f.g.addFootprintStepLocked(&b, item, c.b, nil) {
				t.Fatal("the check refused the other step")
			}
			f.g.markRulesLocked(&b)
			if r := f.conflict(a, b); r != c.want {
				t.Errorf("%q, want %q", r, c.want)
			}
		})
	}
}

// A static that reads a life total or a hand makes a board condition
// see a gain or a draw.
func TestABoardConditionSeesLifeAndCardsThroughAStatic(t *testing.T) {
	f := newCondFixture(t)
	gain := f.mustAccess(t, gainOne)
	draw := f.mustAccess(t, drawOne)
	none := func() layerStatics { return layerStatics{} }
	life := func() layerStatics { return layerStatics{anyLife: true} }
	cards := func() layerStatics { return layerStatics{anyCards: true} }
	if boardConditionSees(&gain, none) || boardConditionSees(&draw, none) {
		t.Error("with no static that reads them, a gain and a draw change no board condition")
	}
	if !boardConditionSees(&gain, life) || boardConditionSees(&draw, life) {
		t.Error("a static that reads a life total sees the gain and not the draw")
	}
	if boardConditionSees(&gain, cards) || !boardConditionSees(&draw, cards) {
		t.Error("a static that reads a hand sees the draw and not the gain")
	}
}

// A player the batch can take out of the game leaves with their
// permanents (CR 800.4a), which a board condition sees.
func TestABoardConditionAsksBesideALossThatCanEliminate(t *testing.T) {
	f := newCondFixture(t)
	f.opp.Life = 1
	a := f.mustAccess(t, cond(FootprintCondBoard, 0), drawOne)
	b := f.mustAccess(t, FootprintStep{Kind: FootprintLoseLifeEachOpponent, N: 1})
	statics := func() layerStatics { return layerStatics{} }
	eliminable := map[uuid.UUID]bool{f.opp.ID: true}
	if w := f.g.accessesConflictLocked(&b, &a, eliminable, statics); w.reason != TriggerOrderReasonElimination {
		t.Errorf("%q, want %q", w.reason, TriggerOrderReasonElimination)
	}
}

func TestAnUnknownConditionIsRefused(t *testing.T) {
	f := newCondFixture(t)
	if _, ok := f.access(t, cond("nonsense", 0), drawOne); ok {
		t.Error("a condition the check does not know must be refused")
	}
}
