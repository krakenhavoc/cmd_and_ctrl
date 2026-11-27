package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 667: Invasion token makers.

const (
	oldB667SaprolingInfestation = "dd257eb3-4112-41e5-9c7d-00d1f683dbdb"
	oldB667PureReflection       = "fee53820-a17c-4f34-9971-581bf8e12fd6"
)

func TestOldB667Registered(t *testing.T) {
	effectstest.StdRegisteredFull(t, oldB667SaprolingInfestation, "Saproling Infestation")
	effectstest.StdRegisteredFull(t, oldB667PureReflection, "Pure Reflection")
}

const oldB667VoyagerDrake = "test-s82-old-b667-kicker-drake"

func oldB667Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

// oldB667Drake is a creature spell of the given mana cost that can be kicked:
// a test spell with Kicker {1}, registered once per test.
func oldB667Drake(t *testing.T, mv string) game.Card {
	t.Helper()
	if _, ok := effects.Lookup(oldB667VoyagerDrake); !ok {
		effectstest.RegisterForTest(t, effects.Spec{OracleID: oldB667VoyagerDrake, Name: "Kicker Drake",
			OptionalCosts: []game.AdditionalCost{effects.Multikicker("{1}", 3)}})
	}
	return game.Card{Name: "Kicker Drake", TypeLine: "Creature — Drake", OracleID: oldB667VoyagerDrake, ManaCost: mv, Power: 3, Toughness: 3}
}

// Saproling Infestation: a kicked spell, anyone's, makes its controller a
// Saproling, once however many times it was kicked; an unkicked spell does not.
func TestOldB667SaprolingInfestation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Saproling Infestation", "Enchantment", oldB667SaprolingInfestation, false)
	if _, err := effectstest.Msc17Cast(t, g, me, oldB667Drake(t, "{3}{U}"), game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB667Count(g, me.ID, "Saproling"); n != 0 {
		t.Fatalf("an unkicked spell made %d Saprolings", n)
	}
	if _, err := effectstest.Msc17Cast(t, g, me, oldB667Drake(t, "{3}{U}"), game.CastSpellParams{OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB667Count(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("a spell kicked twice made %d Saprolings, want one", n)
	}
	// An opponent's kicked spell makes a Saproling for Saproling Infestation's controller.
	effectstest.AdvanceToMainOf(t, g, 1)
	if _, err := effectstest.Msc17Cast(t, g, opp, oldB667Drake(t, "{3}{U}"), game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB667Count(g, me.ID, "Saproling"); n != 2 {
		t.Errorf("my Saprolings %d after an opponent kicked a spell, want 2", n)
	}
	if n := oldB667Count(g, opp.ID, "Saproling"); n != 0 {
		t.Errorf("the caster got %d Saprolings, want none", n)
	}
}

func oldB667Reflections(g *game.Game) map[uuid.UUID][2]int {
	out := map[uuid.UUID][2]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Reflection" {
			out[c.Controller] = [2]int{c.CurrentPower(), c.CurrentToughness()}
		}
	}
	return out
}

// Pure Reflection: each creature spell makes its caster an X/X white
// Reflection, destroying the Reflection before it; a noncreature spell does
// nothing.
func TestOldB667PureReflection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Pure Reflection", "Enchantment", oldB667PureReflection, false)
	if _, err := effectstest.Msc17Cast(t, g, me, oldB667Drake(t, "{3}{U}"), game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := oldB667Reflections(g)
	if len(got) != 1 || got[me.ID] != [2]int{4, 4} {
		t.Fatalf("after a mana value 4 creature spell: %v, want one 4/4 Reflection under my control", got)
	}
	var white bool
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Reflection" {
			white = c.HasColor("W") && c.IsToken()
		}
	}
	if !white {
		t.Error("the Reflection should be a white token")
	}
	// A noncreature spell triggers nothing.
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB667Reflections(g); len(got) != 1 || got[me.ID] != [2]int{4, 4} {
		t.Fatalf("a noncreature spell changed the Reflections: %v", got)
	}
	// The opponent's creature spell replaces it with one of their own.
	effectstest.AdvanceToMainOf(t, g, 1)
	if _, err := effectstest.Msc17Cast(t, g, opp, oldB667Drake(t, "{1}{U}"), game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got = oldB667Reflections(g)
	if len(got) != 1 || got[opp.ID] != [2]int{2, 2} {
		t.Fatalf("after the opponent's mana value 2 spell: %v, want only a 2/2 Reflection under theirs", got)
	}
}
