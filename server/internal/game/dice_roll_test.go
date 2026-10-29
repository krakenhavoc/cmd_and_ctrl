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
