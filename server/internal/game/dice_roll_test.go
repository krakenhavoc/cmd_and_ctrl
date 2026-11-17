package game

import (
	"testing"

	"github.com/google/uuid"
)

// dice_roll_test.go — #4307: the one dice door (dice_roll.go).

func diceRollEvents(g *Game, from int) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == EventRollDie {
				out = append(out, ev.Amount)
			}
		}
	})
	return out
}

func diceIgnorePrompt(g *Game) *PendingChoice {
	var out *PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.PickThen == diceIgnoreKey {
				cp := *c
				out = &cp
			}
		}
	})
	return out
}

// With no replacement, the door rolls exactly what the unreplaceable roll
// does on the same stream, and asks nothing.
func TestRollDiceThenWithNothingToAskIsTheOldRoll(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	control := g.Clone()
	var want []int
	control.WithWriteLock(func() { want, _ = control.RollDiceIgnoringLowestForEffect(RandomDraw{Player: p.ID}, 20, 3, 1) })
	from := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.RollDiceThenForEffect(DiceRoll{Player: p.ID, Sides: 20, N: 3, IgnoreLowest: 1}); err != nil {
			t.Fatalf("roll: %v", err)
		}
	})
	if diceIgnorePrompt(g) != nil {
		t.Fatal("a roll with no replacement asked a question")
	}
	got := diceRollEvents(g, from)
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("announced %v, want %v", got, want)
	}
}

// The shield's choice comes first, then the card's own "ignore the lowest"
// applies to what is left (Berserker's Frenzy under the sticker: three d20,
// one chosen away, then the lower of the other two).
func TestRollDiceThenTheChoiceBeforeTheLowest(t *testing.T) {
	g := newActiveGame(t)
	roller, chooser := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if !g.RollExtraDieNextTimeForEffect(uuid.Nil, chooser.ID, roller.ID, "sticker") {
			t.Fatal("no shield registered")
		}
	})
	from := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.RollDiceThenForEffect(DiceRoll{Player: roller.ID, Sides: 20, N: 2, IgnoreLowest: 1}); err != nil {
			t.Fatalf("roll: %v", err)
		}
	})
	c := diceIgnorePrompt(g)
	if c == nil || c.Chooser != chooser.ID || len(c.PickOptions) != 3 {
		t.Fatalf("the question is %+v, want the chooser's, of three", c)
	}
	rolls := c.dice.Rolls
	if err := g.ResolveOptionPick(c.ID, chooser.ID, 0); err != nil {
		t.Fatalf("answer: %v", err)
	}
	want := max(rolls[1], rolls[2])
	if got := diceRollEvents(g, from); len(got) != 1 || got[0] != want {
		t.Fatalf("announced %v, want [%d] from %v", got, want, rolls)
	}
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if m.Kind == ModRollExtraDieIgnoreOne {
					t.Error("the shield was not spent")
				}
			}
		}
	})
}

func TestRollDiceThenRefusesANonRoll(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	for _, r := range []DiceRoll{
		{Player: p.ID, Sides: 0, N: 1}, {Player: p.ID, Sides: 6, N: 0},
		{Player: p.ID, Sides: 6, N: 2, IgnoreLowest: 2}, {Player: p.ID, Sides: 6, N: 2, IgnoreLowest: -1},
	} {
		g.WithWriteLock(func() {
			if err := g.RollDiceThenForEffect(r); err != ErrInvalidParam {
				t.Errorf("%+v: err %v, want ErrInvalidParam", r, err)
			}
		})
	}
}

// --- #5681: a static "ignore the lowest roll" and a roll's modifier -----

// RollOneMoreAndIgnoreLowest adds a die and one more lowest roll to
// ignore, once per call, asks nobody, and leaves planar dice and every
// other event alone (CR 706.6, 706.7).
func TestRollOneMoreAndIgnoreLowest(t *testing.T) {
	ev := &ReplacementEvent{Kind: RepEventRollDice, RollCount: 1}
	ev.RollOneMoreAndIgnoreLowest()
	ev.RollOneMoreAndIgnoreLowest()
	if ev.RollCount != 3 || ev.rollIgnoreLowest != 2 || len(ev.rollShields) != 0 {
		t.Errorf("after two replacements: count %d, ignore %d, shields %+v", ev.RollCount, ev.rollIgnoreLowest, ev.rollShields)
	}
	planar := &ReplacementEvent{Kind: RepEventRollDice, RollCount: 1, RollPlanar: true}
	planar.RollOneMoreAndIgnoreLowest()
	other := &ReplacementEvent{Kind: RepEventDamage, RollCount: 1}
	other.RollOneMoreAndIgnoreLowest()
	none := &ReplacementEvent{Kind: RepEventRollDice}
	none.RollOneMoreAndIgnoreLowest()
	for name, e := range map[string]*ReplacementEvent{"planar": planar, "damage": other} {
		if e.RollCount != 1 || e.rollIgnoreLowest != 0 {
			t.Errorf("%s: count %d, ignore %d, want it untouched", name, e.RollCount, e.rollIgnoreLowest)
		}
	}
	if none.RollCount != 0 || none.rollIgnoreLowest != 0 {
		t.Error("a roll of no dice was changed")
	}
}

var (
	diceModifierTold     TailOutcome
	diceModifierToldTail = RegisterTail("test/5681-dice-modifier-told", func(_ *Game, o TailOutcome) error {
		diceModifierTold = o
		return nil
	})
)

// A roll's modifier rides the announcement beside the natural result (CR
// 706.2): Amount is the face, Modifier the instruction's, RollResult the
// sum, and the tail is told the natural rolls and the modifier.
func TestRollDiceCarriesItsModifier(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.SetRNGKeyForTest(testRNGKey())
	control := g.Clone()
	var want []int
	control.WithWriteLock(func() { want, _ = control.RollDiceForEffect(RandomDraw{Player: p.ID}, 20, 2) })
	diceModifierTold = TailOutcome{}
	from := len(g.Events)
	g.WithWriteLock(func() {
		r := DiceRoll{Player: p.ID, Sides: 20, N: 2, Modifier: -3, Then: KeyedTail{Ref: diceModifierToldTail, Player: p.ID}}
		if err := g.RollDiceThenForEffect(r); err != nil {
			t.Fatalf("roll: %v", err)
		}
	})
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind != EventRollDie {
				continue
			}
			if ev.Amount != want[n] || ev.Modifier != -3 || ev.RollResult() != want[n]-3 {
				t.Errorf("die %d: natural %d modifier %d result %d, want natural %d modifier -3", n, ev.Amount, ev.Modifier, ev.RollResult(), want[n])
			}
			n++
		}
	})
	if n != 2 {
		t.Fatalf("%d dice announced, want 2", n)
	}
	if len(diceModifierTold.Rolls) != 2 || diceModifierTold.Rolls[0] != want[0] || diceModifierTold.RollModifier != -3 {
		t.Errorf("the tail was told %+v, want the natural rolls %v and modifier -3", diceModifierTold, want)
	}
	if (Event{Kind: EventRollDie, PlanarFace: PlanarFaceChaos, Modifier: 4}).RollResult() != 0 {
		t.Error("a planar die has a numerical result")
	}
}

// A roll waiting on a chooser carries its modifier across a restore: the
// paused roll is written with it (diceRollPauseSnapshot.Modifier).
func TestRollDiceModifierSurvivesTheIgnoreQuestionOnDisk(t *testing.T) {
	g := newActiveGame(t)
	roller := g.Seats[0]
	g.WithWriteLock(func() {
		if !g.RollExtraDieNextTimeForEffect(uuid.Nil, roller.ID, roller.ID, "sticker") {
			t.Fatal("no shield registered")
		}
		r := DiceRoll{Player: roller.ID, Sides: 20, N: 1, Modifier: 7, Then: KeyedTail{Ref: diceModifierToldTail, Player: roller.ID}}
		if err := g.RollDiceThenForEffect(r); err != nil {
			t.Fatalf("roll: %v", err)
		}
	})
	c := diceIgnorePrompt(g)
	if c == nil || c.dice == nil {
		t.Fatal("no roll-to-ignore question")
	}
	back := restoreDiceRollPause(snapshotDiceRollPause(c.dice))
	if back.Roll.Modifier != 7 || back.Roll.Sides != 20 || len(back.Rolls) != 2 {
		t.Fatalf("the paused roll came back as %+v", back)
	}
	diceModifierTold = TailOutcome{}
	from := len(g.Events)
	rolls := c.dice.Rolls
	if err := g.ResolveOptionPick(c.ID, roller.ID, 0); err != nil {
		t.Fatalf("answer: %v", err)
	}
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == EventRollDie && (ev.Amount != rolls[1] || ev.Modifier != 7) {
				t.Errorf("announced natural %d modifier %d, want %d and 7", ev.Amount, ev.Modifier, rolls[1])
			}
		}
	})
	if diceModifierTold.RollModifier != 7 || len(diceModifierTold.Rolls) != 1 {
		t.Errorf("the tail was told %+v", diceModifierTold)
	}
}
