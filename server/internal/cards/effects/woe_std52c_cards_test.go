package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std52c_cards_test.go — the third part of Standard batch 52's tests
// (#4471, #4452): Bitter Chill, Aquatic Alchemist // Bubble Up, Chancellor
// of Tales.

// --- Bitter Chill ----------------------------------------------------------

func TestStd52BitterChillOffersScryAndDrawWhenItGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seedLibrary(me, "A", "B", "C")
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	aura := castFromHand(t, g, game.Card{Name: "Bitter Chill", TypeLine: "Enchantment — Aura", OracleID: std52ChillOracle, ManaCost: "{1}{U}"}, game.CastSpellParams{Targets: cardRefs(theirs)})
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	hand := handSize(me)
	hobDestroy(t, g, aura)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry 1 after paying {1}")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt looks at %d cards, want 1", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d, want %d: scry 1, then draw a card", got, hand+1)
	}
}

func TestStd52BitterChillDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seedLibrary(me, "A", "B", "C")
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	aura := castFromHand(t, g, game.Card{Name: "Bitter Chill", TypeLine: "Enchantment — Aura", OracleID: std52ChillOracle, ManaCost: "{1}{U}"}, game.CastSpellParams{Targets: cardRefs(theirs)})
	passPriorityAroundTable(t, g)
	hand := handSize(me)
	hobDestroy(t, g, aura)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand {
		t.Errorf("hand %d, want %d: nothing was paid", got, hand)
	}
}

// --- Aquatic Alchemist // Bubble Up ----------------------------------------

func TestStd52AquaticAlchemistGrowsOnlyOnTheFirstInstantOrSorceryEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alch := pushCatalogPermanent(g, me.ID, "Aquatic Alchemist", "Creature — Elemental", std52AlchemistOracle, false)
	castFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, alch); p != 1 {
		t.Fatalf("a creature spell pumped the Alchemist to %d", p)
	}
	castFromHand(t, g, game.Card{Name: "Spell One", TypeLine: "Sorcery", ManaCost: "{1}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, alch); p != 3 {
		t.Fatalf("power %d after the first sorcery, want 3", p)
	}
	castFromHand(t, g, game.Card{Name: "Spell Two", TypeLine: "Sorcery", ManaCost: "{1}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, alch); p != 3 {
		t.Errorf("power %d after the second sorcery, want 3", p)
	}
}

func TestStd52BubbleUpPutsAnInstantOrSorceryOnTopAndExilesTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := game.Card{InstanceID: uuid.New(), Name: "Old Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}
	creature := game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(spell)
	me.Graveyard.PushTop(creature)
	c := std52Adventure(me.ID, std52AlchemistOracle, "Aquatic Alchemist", "Creature — Elemental", "{1}{U}", 1, 3, "Bubble Up", "Sorcery — Adventure", "{2}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Strict: true, Targets: cardRefs(creature.InstanceID)}); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(spell.InstanceID)}); err != nil {
		t.Fatalf("cast Bubble Up: %v", err)
	}
	passPriorityAroundTable(t, g)
	if top := libraryTopIDs(me, 1); len(top) != 1 || top[0] != spell.InstanceID {
		t.Errorf("top of library is %v, want Old Spell", top)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("the Adventure card was not exiled")
	}
}

// --- Chancellor of Tales ---------------------------------------------------

func TestStd52ChancellorCopiesAnAdventureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Chancellor of Tales", "Creature — Faerie Advisor", std52ChancellorOracle, false)
	c := std52Adventure(me.ID, std52GiantOracle, "Galvanic Giant", "Creature — Giant Wizard", "{3}{U}", 3, 3, "Storm Reading", "Instant — Adventure", "{5}{U}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	hand := handSize(me)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Storm Reading: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	// The copy and the original each draw four (a discard prompt of two
	// follows each), so at least four more cards than the original alone.
	if got := handSize(me); got < hand-1+8-4 {
		t.Errorf("hand %d, want at least %d with the copy's draw", got, hand-1+8-4)
	}
}
