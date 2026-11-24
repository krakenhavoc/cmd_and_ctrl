package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cast_gate_x_test.go — #6085: the cast gate judges a spell's mana value
// with its announced X (CR 202.3e, CR 601.2b), and CastGateXCeilingLocked
// names the largest X a ManaValueCeiling leaves open.

const gateMVCeilingOracle = "test-gate-mv-ceiling"

// seedManaValueCeiling puts a permanent on the battlefield whose static
// refuses any noncreature spell with mana value greater than `n`.
func seedManaValueCeiling(t *testing.T, g *Game, controller uuid.UUID, n int) {
	t.Helper()
	stubCastRestrictions(t, gateMVCeilingOracle, []CastRestriction{{
		Label: "Test Warden — noncreature spells with mana value greater than n can't be cast.",
		ManaValueCeiling: func(q CastQuery) (int, bool) {
			return n, !q.Card.IsCreature()
		},
	}})
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       "Test Warden",
		TypeLine:   "Enchantment",
		OracleID:   gateMVCeilingOracle,
		Owner:      controller,
		Controller: controller,
	})
}

func TestCastGateJudgesManaValueWithTheAnnouncedX(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		x          int
		refused    bool
	}{
		// {X}{R} at X=1 is mana value 2: under the ceiling.
		{"X at the ceiling", "{X}{R}", 1, false},
		// X=2 makes it 3 (CR 202.3e): refused, though its mana value
		// off the stack is 1.
		{"X above the ceiling", "{X}{R}", 2, true},
		// {X}{X}{R} at X=1 is 3: each {X} counts.
		{"two X slots", "{X}{X}{R}", 1, true},
		{"fixed cost above", "{2}{R}", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			id := uuid.New()
			g.WithWriteLock(func() {
				seedManaValueCeiling(t, g, caster.ID, 2)
				caster.Hand.PushTop(Card{InstanceID: id, Name: "Test Blast", TypeLine: "Instant",
					ManaCost: tc.cost, Owner: caster.ID, Controller: caster.ID})
				for range 6 {
					caster.ManaPool.AddMana(ManaToken{Color: "R"})
				}
			})
			err := g.CastSpell(caster.ID, id, CastSpellParams{XValue: tc.x})
			if refused := errors.Is(err, ErrCantCast); refused != tc.refused {
				t.Fatalf("cast at X=%d: err = %v, want refused %v", tc.x, err, tc.refused)
			}
			if !tc.refused && err != nil {
				t.Fatalf("cast at X=%d: %v", tc.x, err)
			}
			if tc.refused && !caster.Hand.Contains(id) {
				t.Error("a refused cast left the hand")
			}
		})
	}
}

func TestCastGateXCeilingFollowsTheManaValueCeiling(t *testing.T) {
	g := newRestorableGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { seedManaValueCeiling(t, g, caster.ID, 5) })
	for _, tc := range []struct {
		name, typeLine, cost string
		want                 int
		ok                   bool
	}{
		{"one X slot", "Sorcery", "{X}{R}", 4, true},
		{"two X slots", "Sorcery", "{X}{X}{R}", 2, true},
		{"ceiling below the base", "Sorcery", "{X}{6}", 0, true},
		{"no X slot", "Sorcery", "{4}{R}", 0, false},
		{"restriction does not apply", "Creature — Hydra", "{X}{G}", 0, false},
	} {
		card := Card{InstanceID: uuid.New(), Name: "Test", TypeLine: tc.typeLine, ManaCost: tc.cost,
			Owner: caster.ID, Controller: caster.ID}
		var got int
		var ok bool
		g.ReadSnapshot(func() { got, ok = g.CastGateXCeilingLocked(caster.ID, card, ZoneHand) })
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: ceiling = (%d, %v), want (%d, %v)", tc.name, got, ok, tc.want, tc.ok)
		}
	}
	// And the auto-tap preview's ceiling folds it in.
	card := Card{InstanceID: uuid.New(), Name: "Test", TypeLine: "Sorcery", ManaCost: "{X}{R}",
		Owner: caster.ID, Controller: caster.ID}
	g.WithWriteLock(func() { caster.Hand.PushTop(card) })
	if got, ok := g.SpellXCeilingForCast(caster.ID, card, CastSpellParams{}); !ok || got != 4 {
		t.Errorf("SpellXCeilingForCast = (%d, %v), want (4, true)", got, ok)
	}
}
