package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// answers_saves_test.go is #2939 gap 3: a row on another permanent that
// saves this one (Selfless Spirit).

// savior is a 2/1 whose only row is "Sacrifice this creature: creatures
// you control gain indestructible until end of turn", declared as
// Selfless Spirit declares it.
func savior(matches string, partial bool, answers ...string) protocol.CardView {
	c := creature(cardID(1), 0, "Selfless Spirit", 2, 1)
	c.ManaCost = "{1}{W}"
	a := append([]string(nil), answers...)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "Sacrifice this creature: Creatures you control gain indestructible until end of turn.",
		SacrificeSelf: true,
		Purpose:       &protocol.PurposeView{Answers: &a, Saves: &protocol.SavesView{Matches: matches, Partial: partial}},
	}}
	return c
}

// ogre is a 4/4 worth saving, with a mana value of four.
func ogre(id string) protocol.CardView {
	c := creature(id, 0, "Ogre", 4, 4)
	c.ManaCost = "{3}{R}"
	return c
}

// titan is a 7/7, worth a Spirit even at the odds a Murder destroys it.
func titan(id string) protocol.CardView {
	c := creature(id, 0, "Titan", 7, 7)
	c.ManaCost = "{5}{R}{R}"
	c.Abilities = []string{"flying", "lifelink"}
	return c
}

// bigWrath is an opponent's declared sweep of the creatures with mana
// value four or more, which a two-mana Spirit lives through.
func bigWrath(controller int, id string) (protocol.CardView, protocol.StackItemView) {
	return bigSweep(controller, id, "destroy")
}

// bigSweep is bigWrath that takes the creatures by how.
func bigSweep(controller int, id, how string) (protocol.CardView, protocol.StackItemView) {
	c := sorcery(id, controller, "Austere Command", "{4}{W}{W}",
		&protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures_mana_value_4_or_greater", How: how}})
	return c, protocol.StackItemView{ID: "item-" + id, Kind: "spell", Controller: seatID(controller).String(), SourceCardID: id}
}

// noSaveOthers is the heuristic with only #2939 gap 3's switch off.
func noSaveOthers() *heuristic.Policy {
	c := heuristic.DefaultConfig()
	c.PriceSaveOthers = false
	return heuristic.NewWithConfig(c)
}

const sacLabel = "Sacrifice the Spirit"

// The Spirit is sacrificed to save the Ogres from a sweep it lives
// through, or from a Murder aimed at one of them, and not when nothing
// threatens them. A Wrath of God it would die to anyway already got it
// sacrificed, free, before this change; it is the Spirit that survives
// the sweep, or the removal aimed past it, that the row was never
// priced for.
func TestSelflessSpiritSavesTheOgres(t *testing.T) {
	big, bi := bigWrath(1, cardID(9))
	div, di := divination(1, cardID(9))
	murder, mi := removal(1, cardID(9), cardID(2))
	two := func(answers string) []protocol.CardView {
		return []protocol.CardView{savior("creatures", false, answers), ogre(cardID(2)), ogre(cardID(3))}
	}
	for _, tc := range []struct {
		name  string
		cards []protocol.CardView
		stack protocol.CardView
		item  protocol.StackItemView
		want  string
	}{
		{"a sweep of the big ones, two ogres", two("survive"), big, bi, sacLabel},
		{"a sweep of the big ones, a lone spirit", []protocol.CardView{savior("creatures", false, "survive")}, big, bi, "Pass priority"},
		{"divination threatens nothing", two("survive"), div, di, "Pass priority"},
		{"murder aimed at a titan", []protocol.CardView{savior("creatures", false, "survive"), titan(cardID(2))}, murder, mi, sacLabel},
		{"murder aimed at an ogre is not worth the Spirit", two("survive")[:2], murder, mi, "Pass priority"},
		{"hexproof does nothing against a sweep", two("protect"), big, bi, "Pass priority"},
		{"hexproof answers a murder", two("protect")[:2], murder, mi, sacLabel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := input(0, respond(tc.cards, tc.stack, tc.item), passMove(0), activate(t, 0, cardID(1), sacLabel, nil))
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
				t.Errorf("the heuristic chose %q, want %q", got, tc.want)
			}
			if tc.want != "Pass priority" {
				if got := chose(t, in, decide(t, noSaveOthers(), in)); got != "Pass priority" {
					t.Errorf("without PriceSaveOthers it chose %q, want the old pass", got)
				}
				if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
					t.Errorf("the baseline chose %q, want a pass", got)
				}
			}
		})
	}
}

// The Spirit is not sacrificed for a creature it is not worth: a Murder
// at a Bear costs the bot more than the Bear saves.
func TestSelflessSpiritIsNotSacrificedForABear(t *testing.T) {
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	murder, mi := removal(1, cardID(9), cardID(2))
	in := input(0, respond([]protocol.CardView{savior("creatures", false, "survive"), bear}, murder, mi), passMove(0), activate(t, 0, cardID(1), sacLabel, nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// Indestructible is no answer to a sweep that exiles, makes the bot
// sacrifice, bounces or shrinks, so the Spirit is kept (Farewell).
func TestSelflessSpiritIsNoAnswerToAnExilingSweep(t *testing.T) {
	for _, how := range []string{"exile", "sacrifice", "bounce", "minus"} {
		sweep, si := bigSweep(1, cardID(9), how)
		cards := []protocol.CardView{savior("creatures", false, "survive"), ogre(cardID(2)), ogre(cardID(3))}
		in := input(0, respond(cards, sweep, si), passMove(0), activate(t, 0, cardID(1), sacLabel, nil))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("a sweep that takes the creatures by %s: the heuristic chose %q, want a pass", how, got)
		}
	}
}

// A targeted ability is presumed removal, as any targeted item is, but
// not by a row that costs the bot a card: a planeswalker's "target
// creature can't be blocked" is no reason to sacrifice the Spirit.
func TestSelflessSpiritIgnoresATargetedAbility(t *testing.T) {
	walker := spell(cardID(9), 1, "Planeswalker", "{3}{W}")
	item := protocol.StackItemView{
		ID: "item-ability", Kind: "activated", Controller: seatID(1).String(), SourceCardID: cardID(9),
		Targets: []protocol.TargetRefView{{Kind: "card", ID: cardID(2)}},
	}
	cards := []protocol.CardView{savior("creatures", false, "survive"), ogre(cardID(2))}
	in := input(0, respond(cards, walker, item), passMove(0), activate(t, 0, cardID(1), sacLabel, nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// A second Spirit does not answer a threat the first already has, whose
// ability is on the stack above it from a source that has left the
// battlefield.
func TestSecondSpiritDoesNotAnswerTheSameSweep(t *testing.T) {
	big, bi := bigWrath(1, cardID(9))
	first := protocol.StackItemView{ID: "first", Kind: "ability", Controller: seatID(0).String(), SourceCardID: cardID(7)}
	cards := []protocol.CardView{savior("creatures", false, "survive"), ogre(cardID(2)), ogre(cardID(3))}
	in := input(0, respond(cards, big, bi, first), passMove(0), activate(t, 0, cardID(1), sacLabel, nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// A row that covers only part of the class is worth half as much: Ainok
// Strike Leader saves tokens, not every creature.
func TestPartialSavesPriceHalf(t *testing.T) {
	pol := heuristic.New()
	big, bi := bigWrath(1, cardID(9))
	price := func(partial bool) float64 {
		cards := []protocol.CardView{savior("creatures", partial, "survive"), ogre(cardID(2)), ogre(cardID(3)), ogre(cardID(4))}
		return rankValue(t, pol, input(0, respond(cards, big, bi), passMove(0), activate(t, 0, cardID(1), sacLabel, nil)), sacLabel)
	}
	full, half := price(false), price(true)
	if half >= full {
		t.Errorf("a partial save priced %.3f, want below the full class's %.3f", half, full)
	}
}
