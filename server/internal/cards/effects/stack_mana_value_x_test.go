package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_mana_value_x_test.go — #788. CR 202.3e: while a spell is on
// the stack, the {X} in its mana cost counts as the value chosen for
// it. Every card here used to read the printed cost, where X is zero.
// They now read game.(*Game).ManaValueForEffect. The engine table for
// that read is game/mana_value_for_effect_test.go. These tests check
// each card that used the printed cost.

// seedLibraryOfCost replaces the active seat's library with `n`
// instants that all cost `manaCost`, so the cascade limit alone
// decides whether anything is a hit.
func seedLibraryOfCost(g *game.Game, p *game.Player, n int, manaCost string) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		c := game.NewCard("Library Spell", p.ID)
		c.TypeLine = "Instant"
		c.ManaCost = manaCost
		p.Library.PushTop(c)
	}
}

// The shared mana-value target predicates all make the same read.
// A {X}{R} spell cast with X=3 is mana value 4 on the stack, and its
// printed cost is 1.
func TestManaValuePredicatesCountXOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castXSpell(t, g, "Red X Spell", "Sorcery", "", "{X}{R}", 3, nil)
	spell, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the spell is not on the stack")
	}
	for _, row := range []struct {
		name string
		pred CardPredicate
		want bool
	}{
		{"ManaValueLE(3)", ManaValueLE(3), false},
		{"ManaValueLE(4)", ManaValueLE(4), true},
		{"ManaValueGE(4)", ManaValueGE(4), true},
		{"ManaValueGE(5)", ManaValueGE(5), false},
		{"b03ManaValueIs(1)", b03ManaValueIs(1), false},
		{"b03ManaValueIs(4)", b03ManaValueIs(4), true},
	} {
		if got := row.pred(g, me.ID, spell); got != row.want {
			t.Errorf("%s on {X}{R} with X=3 = %v, want %v", row.name, got, row.want)
		}
	}
}

// Mental Misstep reads its target's mana value on the stack. Reading
// the printed cost would let it counter a {X}{U} spell cast with X=1
// (mana value 2), which is stronger than printed.
func TestMentalMisstepCountsXOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	two := castXSpell(t, g, "Blue X Artifact", "Artifact", "", "{X}{U}", 1, nil)
	if !b03CastRefused(t, g, "Mental Misstep", "Instant", b03MentalMisstepOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: two}},
	}) {
		t.Error("{X}{U} cast with X=1 is mana value 2 and is not a legal target")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Fatal("the artifact that was not countered should have resolved")
	}

	// A countered artifact goes to the graveyard. One that resolves
	// goes to the battlefield. Both of these are mana value 1.
	for _, row := range []struct {
		name string
		cost string
		x    int
	}{
		{"{X}{U} cast with X=0", "{X}{U}", 0},
		{"{X} cast with X=1", "{X}", 1},
	} {
		one := castXSpell(t, g, "X Artifact", "Artifact", "", row.cost, row.x, nil)
		castCatalogSpell(t, g, "Mental Misstep", "Instant", b03MentalMisstepOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: one}})
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(one) || !me.Graveyard.Contains(one) {
			t.Errorf("%s is mana value 1 and should be countered", row.name)
		}
	}
}

// Imoti's "mana value 6 or greater" reads the spell on the stack. A
// {X}{G} spell cast with X=4 is mana value 5, and one cast with X=5
// is 6.
func TestImotiCountsXOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 6)
	pushCatalogPermanent(g, me.ID, "Imoti, Celebrant of Bounty", "Legendary Creature — Snake Druid", imotiOracle, false)

	castXSpell(t, g, "Green X Spell", "Sorcery", "", "{X}{G}", 4, nil)
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("{X}{G} with X=4 (mana value 5) got %d cascade offers, want 0", got)
	}

	castXSpell(t, g, "Green X Spell", "Sorcery", "", "{X}{G}", 5, nil)
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Errorf("{X}{G} with X=5 (mana value 6) got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

// The limit a cascade records is the spell's mana value on the stack
// (CR 702.85a), X included. That holds whether the spell has cascade
// itself (Cascade, on Bloodbraid Elf) or is given it (GrantsCascade,
// on Maelstrom Nexus). Every card in the library is mana value 5, so
// there is a hit only when the cascading spell is 6 or more.
func TestCascadeLimitCountsXOnTheStack(t *testing.T) {
	rows := []struct {
		name   string
		oracle string // the cascader's own oracle; empty when the Nexus grants cascade
		cost   string
		x      int
		offers int
	}{
		{"granted, {X}{G} with X=3 is 4: a five-drop is too big", "", "{X}{G}", 3, 0},
		{"granted, {X}{G} with X=5 is 6: a five-drop is a hit", "", "{X}{G}", 5, 1},
		{"own keyword, {X}{R}{G} with X=2 is 4: a five-drop is too big", bloodbraidElfOracle, "{X}{R}{G}", 2, 0},
		{"own keyword, {X}{R}{G} with X=4 is 6: a five-drop is a hit", bloodbraidElfOracle, "{X}{R}{G}", 4, 1},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibraryOfCost(g, me, 6, "{4}{U}")
			if row.oracle == "" {
				pushCatalogPermanent(g, me.ID, "Maelstrom Nexus", "Enchantment", maelstromNexusOracle, false)
			}
			castXSpell(t, g, "X Cascader", "Sorcery", row.oracle, row.cost, row.x, nil)
			passPriorityAroundTable(t, g)
			if got := countMayCastPrompts(g); got != row.offers {
				t.Errorf("cascade offers = %d, want %d", got, row.offers)
			}
			answerAllMayCast(t, g, me.ID, false)
		})
	}
}

// Sanctum of Ugin reads the cast spell's mana value on the stack. A
// Walking Ballista cast with X=3 is mana value 6, and one cast with
// X=4 is 8.
func TestB27SanctumOfUginCountsXOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Sanctum of Ugin", "Land", b27SanctumOfUginOracle, 0, 0)
	seedSearchLibrary(me,
		game.Card{Name: "Ulamog", TypeLine: "Legendary Creature — Eldrazi", ManaCost: "{10}"},
	)
	castXSpell(t, g, "Walking Ballista", "Artifact Creature — Construct", "", "{X}{X}", 3, nil)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("X=3 is mana value 6: the Sanctum does not trigger")
	}
	castXSpell(t, g, "Walking Ballista", "Artifact Creature — Construct", "", "{X}{X}", 4, nil)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("X=4 is mana value 8: the Sanctum triggers")
	}
}
