package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const darkstarAugurOracle = "67d2a021-a042-4e5f-b6e4-39cb35514794"

// castDarkstarAugur casts the Augur out of the active seat's hand
// with the named optional costs claimed. It seeds the printed P/T
// itself, because the 1/1 of the offspring token is only an assertion
// against a body that is not 1/1 to begin with.
func castDarkstarAugur(t *testing.T, g *game.Game, optional []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Darkstar Augur",
		TypeLine:   "Creature — Bat Warlock",
		OracleID:   darkstarAugurOracle,
		ManaCost:   "{2}{B}",
		Power:      2,
		Toughness:  3,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell Darkstar Augur: %v", err)
	}
	return id
}

// TestDarkstarAugurOffspringMakesOneOneOrNothing is the keyword's
// whole contract in one test, and the pair is the assertion: the
// token exists only when the cost was paid, it is a 1/1 copy of the
// body rather than a second 2/3, and — the part a recursive keyword
// gets wrong — the TOKEN does not trigger offspring again.
func TestDarkstarAugurOffspringMakesOneOneOrNothing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unpaid", nil, 1},
		{"offspring paid", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			cast := castDarkstarAugur(t, g, tc.optional)
			passPriorityAroundTable(t, g)

			var copies []game.Card
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Darkstar Augur" {
					copies = append(copies, c)
				}
			}
			// An offspring token that triggered offspring itself would
			// make this grow without bound, so the count IS the
			// recursion assertion.
			if len(copies) != tc.want {
				t.Fatalf("Darkstar Augurs on the battlefield: got %d, want %d", len(copies), tc.want)
			}
			for _, c := range copies {
				if c.InstanceID == cast {
					if c.Power != 2 || c.Toughness != 3 {
						t.Errorf("the cast creature is a %d/%d, want 2/3", c.Power, c.Toughness)
					}
					continue
				}
				if !c.IsToken() {
					t.Errorf("the offspring copy is not a token: %q", c.TypeLine)
				}
				if c.Power != 1 || c.Toughness != 1 {
					t.Errorf("the offspring token is a %d/%d, want 1/1", c.Power, c.Toughness)
				}
				// A real copy, so it carries the oracle ID and with it
				// every catalogued ability — the flying included.
				if c.OracleID != darkstarAugurOracle {
					t.Errorf("the offspring token is not a copy: oracle %q", c.OracleID)
				}
				if got := effectiveAbilities(t, g, c.InstanceID); !slices.Contains(got, "flying") {
					t.Errorf("the offspring token does not fly: %v", got)
				}
			}
		})
	}
}

// TestDarkstarAugurFlies is the printed keyword, read off the
// layered characteristics the combat engine reads.
func TestDarkstarAugurFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushPermanentForTest(g, me.ID, "Darkstar Augur", darkstarAugurOracle, "Creature — Bat Warlock")
	if got := effectiveAbilities(t, g, id); !slices.Contains(got, "flying") {
		t.Errorf("Darkstar Augur should fly: %v", got)
	}
}

// TestDarkstarAugurUpkeepFlipCostsItsManaValue is the Dark Confidant
// half: the top card is revealed, it lands in hand, and the life lost
// is its mana value — which is zero for a land, and zero is not a
// special case but the ordinary answer.
func TestDarkstarAugurUpkeepFlipCostsItsManaValue(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		manaCost string
		want     int
	}{
		{"a land costs nothing", "Land", "", 0},
		{"a three-drop costs three", "Creature — Bear", "{2}{B}", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			pushPermanentForTest(g, me.ID, "Darkstar Augur", darkstarAugurOracle, "Creature — Bat Warlock")
			flip := uuid.New()
			me.Library.PushTop(game.Card{
				InstanceID: flip,
				Name:       "Flipped Card",
				TypeLine:   tc.typeLine,
				ManaCost:   tc.manaCost,
				Owner:      me.ID,
				Controller: me.ID,
			})
			life, hand := me.Life, me.Hand.Size()

			advanceToUpkeepOf(t, g, 1)
			passPriorityAroundTable(t, g)

			if !me.Hand.Contains(flip) {
				t.Fatalf("the revealed card did not reach the hand")
			}
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand %d -> %d, want exactly one card", hand, me.Hand.Size())
			}
			if want := life - tc.want; me.Life != want {
				t.Errorf("life %d -> %d, want %d", life, me.Life, want)
			}
		})
	}
}
