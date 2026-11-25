package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_be_countered_by_color_test.go — #6263, Autumn's Veil: "Spells you
// control can't be countered by blue or black spells this turn, and
// creatures you control can't be the targets of blue or black spells this
// turn." CR 701.6a: to counter is to remove from the stack; the grant
// stops only a counter made by a blue or black SPELL.

// resolveAs parks a spell of `colors` (or an ability, when kind is not a
// spell) in the resolving slot, as the resolution of a counterspell does.
func resolveAs(g *Game, controller uuid.UUID, kind StackItemKind, colors ...string) {
	g.WithWriteLock(func() {
		card := Card{InstanceID: uuid.New(), Name: "Counterer", TypeLine: "Instant", Colors: colors, Owner: controller, Controller: controller}
		item := &StackItem{ID: card.InstanceID, Kind: kind, Controller: controller}
		if kind == StackItemSpell {
			g.beginResolvingSpellLocked(item, card)
		} else {
			g.beginResolvingLocked(item)
		}
	})
}

func stillOnStack(g *Game, id uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() { ok = g.Stack.Contains(id) })
	return ok
}

func autumnsVeilShield() CounterShieldGrant {
	return CounterShieldGrant{Text: "Spells you control can't be countered by blue or black spells this turn.", ByColors: "UB"}
}

// A blue counterspell can't counter the spell; a black one can't either.
func TestByColorShieldStopsABlueOrBlackCounterspell(t *testing.T) {
	for _, colors := range [][]string{{"U"}, {"B"}, {"U", "R"}} {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spell := pushSpellFor(t, g, "Sorcery", me, me)
		grantShield(g, me, autumnsVeilShield(), "Autumn's Veil")
		resolveAs(g, opp.ID, StackItemSpell, colors...)
		var err error
		g.WithWriteLock(func() { err = g.CounterTargetForEffect(spell) })
		if err != nil {
			t.Fatalf("%v counter: %v", colors, err)
		}
		if !stillOnStack(g, spell) {
			t.Errorf("a %v spell countered the spell under Autumn's Veil", colors)
		}
	}
}

// A colourless spell, a red spell and an ability still counter it.
func TestByColorShieldLetsOtherCountersThrough(t *testing.T) {
	cases := []struct {
		name   string
		kind   StackItemKind
		colors []string
	}{
		{"colourless spell", StackItemSpell, nil},
		{"red spell", StackItemSpell, []string{"R"}},
		{"blue ability", StackItemTriggered, []string{"U"}},
	}
	for _, c := range cases {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spell := pushSpellFor(t, g, "Sorcery", me, me)
		grantShield(g, me, autumnsVeilShield(), "Autumn's Veil")
		resolveAs(g, opp.ID, c.kind, c.colors...)
		g.WithWriteLock(func() { _ = g.CounterTargetForEffect(spell) })
		if stillOnStack(g, spell) {
			t.Errorf("%s: the spell was not countered", c.name)
		}
	}
}

// With no counter being made, the spell is not "can't be countered":
// the stack chip does not claim more than the grant says.
func TestByColorShieldIsNotABlanketShield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Sorcery", me, me)
	grantShield(g, me, autumnsVeilShield(), "Autumn's Veil")
	if cantBeCountered(g, spell) {
		t.Error("the colour-qualified grant reads as can't be countered by anything")
	}
}

// The targeting half: a blue or black spell can't target a creature the
// Veil's controller controls, even one that arrives later (CR 611.2c),
// whoever casts it; a red spell and a blue ability still can.
func TestColorSpellTargetingBar(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.New(), ScopeYourCreatures, me.ID,
			[]Mod{CantBeTargetedBySpellsOfColorsMod("U", "B")}, g.UntilEndOfTurnDuration(), "Autumn's Veil")
	})
	mine := pushBattlefieldForTest(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := pushBattlefieldForTest(g, opp.ID, "Their Bear", "Creature — Bear", "")
	blue := &Card{Name: "Blue Spell", TypeLine: "Instant", Colors: []string{"U"}}
	red := &Card{Name: "Red Spell", TypeLine: "Instant", Colors: []string{"R"}}
	g.WithWriteLock(func() {
		m, th := g.findBattlefieldCardLocked(mine), g.findBattlefieldCardLocked(theirs)
		if canBeTargetedBy(g, m, ZoneBattlefield, SourceCast(opp.ID, blue)) {
			t.Error("an opponent's blue spell targets my creature")
		}
		if canBeTargetedBy(g, m, ZoneBattlefield, SourceCast(me.ID, blue)) {
			t.Error("my own blue spell targets my creature")
		}
		if !canBeTargetedBy(g, m, ZoneBattlefield, SourceCast(opp.ID, red)) {
			t.Error("a red spell can't target my creature")
		}
		if !canBeTargetedBy(g, m, ZoneBattlefield, SourceObject(opp.ID, blue)) {
			t.Error("a blue ability's source can't target my creature")
		}
		if !canBeTargetedBy(g, th, ZoneBattlefield, SourceCast(opp.ID, blue)) {
			t.Error("the bar reached an opponent's creature")
		}
	})
}

func TestColorSpellTargetingBarModProblem(t *testing.T) {
	if p := colorSpellTargetingBarModProblem(CantBeTargetedBySpellsOfColorsMod("U", "B")); p != "" {
		t.Errorf("sound mod reported %q", p)
	}
	if p := colorSpellTargetingBarModProblem(CantBeTargetedBySpellsOfColorsMod()); p == "" {
		t.Error("a mod naming no colour passed")
	}
	if p := colorSpellTargetingBarModProblem(CantBeTargetedBySpellsOfColorsMod("X")); p == "" {
		t.Error("a mod naming a non-colour passed")
	}
}
