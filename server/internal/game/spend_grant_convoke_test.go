package game

import (
	"testing"

	"github.com/google/uuid"
)

// spend_grant_convoke_test.go — #1928. A cast permission's "spend mana
// as though it were mana of any color" (Breeches, Hostage Taker) changes
// how MANA may pay a symbol (CR 609.4b). Convoke is not mana
// (CR 702.51a): a creature pays {1} or one mana of its own colour, so a
// stolen convoke spell's coloured symbols still need creatures of that
// colour, or mana, which the grant does widen.

const stolenConvokeOracle = "test-1928-stolen-convoke"

// stolenConvokeFixture exiles a {2}{R}{R} convoke spell off the
// opponent's library under `perm`, in the thief's main phase.
func stolenConvokeFixture(t *testing.T, perm CastPermission) (*Game, *Player, uuid.UUID) {
	t.Helper()
	prev := CatalogTapPermanentsCost
	CatalogTapPermanentsCost = func(id string) *TapPermanentsCost {
		if id == stolenConvokeOracle {
			return convokeForTest()
		}
		return nil
	}
	t.Cleanup(func() { CatalogTapPermanentsCost = prev })

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	c := NewCard("Stolen Stoke", opp.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{2}{R}{R}"
	c.OracleID = stolenConvokeOracle
	opp.Library.PushTop(c)
	return g, me, impulseExile(t, g, opp, me, perm)
}

func convokeCreature(g *Game, owner *Player, colour string) uuid.UUID {
	c := Card{Name: "Convoker", TypeLine: "Creature — Bear", Owner: owner.ID, Controller: owner.ID,
		Power: 1, Toughness: 1}
	if colour != "" {
		c.Colors = []string{colour}
	}
	return pushTypedTestCard(g, c)
}

var anyColourGrants = []struct {
	name string
	perm CastPermission
}{
	{"any color", CastPermission{AnyColor: true, Duration: WhileInZoneDuration()}},
	{"any type", CastPermission{AnyType: true, Duration: WhileInZoneDuration()}},
}

// Four green creatures cannot pay {2}{R}{R}: two of them pay the
// generic and the {R}{R} are still owed, because a creature is not mana.
func TestStolenConvokeOffColourCreaturesCannotPayColouredSymbols(t *testing.T) {
	for _, tc := range anyColourGrants {
		t.Run(tc.name, func(t *testing.T) {
			g, me, spell := stolenConvokeFixture(t, tc.perm)
			var taps []uuid.UUID
			for i := 0; i < 4; i++ {
				taps = append(taps, convokeCreature(g, me, "G"))
			}
			if err := g.CastSpell(me.ID, spell, CastSpellParams{FromZone: "exile", Strict: true, TapIDs: taps}); err == nil {
				t.Fatal("four green creatures paid {2}{R}{R} under an any-colour grant")
			}
			if g.Stack.Contains(spell) {
				t.Error("the refused cast is on the stack")
			}
		})
	}
}

// What the grant does widen is MANA: two off-colour creatures pay the
// generic and two green mana pay the {R}{R} as though they were red.
func TestStolenConvokeOffColourManaPaysTheColouredSymbols(t *testing.T) {
	for _, tc := range anyColourGrants {
		t.Run(tc.name, func(t *testing.T) {
			g, me, spell := stolenConvokeFixture(t, tc.perm)
			taps := []uuid.UUID{convokeCreature(g, me, "G"), convokeCreature(g, me, "G")}
			me.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "G"})
			if err := g.CastSpell(me.ID, spell, CastSpellParams{FromZone: "exile", Strict: true, TapIDs: taps}); err != nil {
				t.Fatalf("two creatures plus two green mana: %v", err)
			}
			if !g.Stack.Contains(spell) || len(me.ManaPool) != 0 {
				t.Errorf("on stack %v, pool %v: the cast should have spent the mana", g.Stack.Contains(spell), me.ManaPool)
			}
		})
	}
}

// An on-colour creature still pays a coloured symbol: two red and two
// green creatures cast the spell with no mana at all.
func TestStolenConvokeOnColourCreaturesPayColouredSymbols(t *testing.T) {
	for _, tc := range anyColourGrants {
		t.Run(tc.name, func(t *testing.T) {
			g, me, spell := stolenConvokeFixture(t, tc.perm)
			taps := []uuid.UUID{
				convokeCreature(g, me, "R"), convokeCreature(g, me, "R"),
				convokeCreature(g, me, "G"), convokeCreature(g, me, "G"),
			}
			if err := g.CastSpell(me.ID, spell, CastSpellParams{FromZone: "exile", Strict: true, TapIDs: taps}); err != nil {
				t.Fatalf("two red and two green creatures: %v", err)
			}
			if !g.Stack.Contains(spell) {
				t.Error("the spell is not on the stack")
			}
		})
	}
}

// The grant changes how the cost is paid, not the cost (CR 609.4b), so
// the price shown is the printed one.
func TestSpendGrantLeavesTheShownPricePrinted(t *testing.T) {
	for _, tc := range anyColourGrants {
		t.Run(tc.name, func(t *testing.T) {
			g, me, spell := stolenConvokeFixture(t, tc.perm)
			card, ok := g.LookupCardForEffect(spell)
			if !ok {
				t.Fatal("spell not found")
			}
			price, err := g.PriceCast(me.ID, card, CastSpellParams{FromZone: "exile"})
			if err != nil {
				t.Fatal(err)
			}
			if got := price.Total.String(); got != "{2}{R}{R}" {
				t.Errorf("shown price = %s, want the printed {2}{R}{R}", got)
			}
		})
	}
}
