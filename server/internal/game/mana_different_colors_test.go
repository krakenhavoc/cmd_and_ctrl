package game

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
)

// mana_different_colors_test.go — #2558: "Add two mana of different
// colors" (Firemind Vessel). One slot, "{W|U|B|R|G:2}", whose two mana
// must be two DIFFERENT colours: the prompt asks twice and refuses a
// repeat, a colour named up front is refused if repeated, and the
// auto-tapper books a pair that pays the cost and never a pair of one
// colour. The restore and the cards are tested with the catalog, in
// cards/effects/different_colors_cards_test.go.

const differentTwo = "{W|U|B|R|G:2}"

func differentShape() []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: differentTwo, Label: "Add two mana of different colors"}}
}

func pushDifferentSource(g *Game, owner *Player) uuid.UUID {
	return pushIntrinsicPermanent(g, owner, "Firemind Vessel", "Artifact", differentShape(), nil)
}

// openManaPicks is every PendingChoiceMana in the queue.
func openManaPicks(g *Game) []*PendingChoice {
	var out []*PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMana {
			out = append(out, c)
		}
	}
	return out
}

func sortedPool(p *Player) []string {
	out := make([]string, 0, len(p.ManaPool))
	for _, tok := range p.ManaPool {
		out = append(out, tok.Color)
	}
	sort.Strings(out)
	return out
}

func TestParseProducedManaDifferentColors(t *testing.T) {
	slots, err := ParseProducedMana(differentTwo)
	if err != nil {
		t.Fatalf("ParseProducedMana(%q): %v", differentTwo, err)
	}
	if len(slots) != 1 || slots[0].Distinct != 2 || !reflect.DeepEqual(slots[0].Options, AllColors) {
		t.Fatalf("slots = %+v, want one slot of all five colours, Distinct 2", slots)
	}
	if !slots[0].DifferentColors() || slots[0].OneColorAmounts() {
		t.Errorf("a different-colours slot is not a one-colour pick: %+v", slots[0])
	}
	// Exactly N options is no choice: one fixed slot per colour.
	fixed, err := ParseProducedMana("{W|U:2}")
	if err != nil {
		t.Fatalf("ParseProducedMana({W|U:2}): %v", err)
	}
	if len(fixed) != 2 || fixed[0].Distinct != 0 || fixed[0].Options[0] != "W" || fixed[1].Options[0] != "U" {
		t.Errorf("{W|U:2} = %+v, want {W}{U}", fixed)
	}
	for _, bad := range []string{
		"{W|U|B:1}",   // one is an ordinary pick
		"{W|U|B:}",    // no count
		"{W|U:3}",     // more colours than options
		"{W|U|C:2}",   // colourless is not a colour
		"{W2|U2|B:2}", // amounts and different colours do not mix
	} {
		if _, err := ParseProducedMana(bad); err == nil {
			t.Errorf("ParseProducedMana(%q) accepted a malformed different-colours slot", bad)
		}
	}
}

// Two different colours are required: the prompt asks for a first
// colour, then a second with the first struck out, and adds nothing
// until the second is named — then both, together.
func TestDifferentColorsPromptAsksForTwoDifferentColors(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	vessel := pushDifferentSource(g, me)

	if err := g.ActivateManaAbility(me.ID, vessel, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := openManaPicks(g)
	if len(picks) != 1 {
		t.Fatalf("queued %d mana picks, want ONE pick of two colours", len(picks))
	}
	first := picks[0]
	if first.ManaDifferent != 2 || len(first.ColorOptions) != 5 {
		t.Fatalf("first pick = %+v, want two different colours from five", first)
	}
	firstID := first.ID
	if err := g.ResolveManaChoice(firstID, me.ID, "U"); err != nil {
		t.Fatalf("answer U: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Fatalf("pool = %v after the first colour; the mana arrives together", me.ManaPool)
	}
	picks = openManaPicks(g)
	if len(picks) != 1 {
		t.Fatalf("after the first answer %d picks are open, want the same one", len(picks))
	}
	second := picks[0]
	if second.ID == firstID {
		t.Error("the second question kept the first one's ID; a seat could read it as answered")
	}
	if containsColor(second.ColorOptions, "U") || len(second.ColorOptions) != 4 {
		t.Errorf("second pick offers %v, want the four colours other than blue", second.ColorOptions)
	}
	if !reflect.DeepEqual(second.ManaChosen, []string{"U"}) {
		t.Errorf("ManaChosen = %v, want [U]", second.ManaChosen)
	}
	// The old ID is gone: answering it again is not a second pick.
	if err := g.ResolveManaChoice(firstID, me.ID, "W"); !errors.Is(err, ErrPendingChoiceNotFound) {
		t.Errorf("answering the first question's ID again = %v, want ErrPendingChoiceNotFound", err)
	}
	if err := g.ResolveManaChoice(second.ID, me.ID, "G"); err != nil {
		t.Fatalf("answer G: %v", err)
	}
	if got := sortedPool(me); !reflect.DeepEqual(got, []string{"G", "U"}) {
		t.Errorf("pool = %v, want [G U]", got)
	}
	if left := openManaPicks(g); len(left) != 0 {
		t.Errorf("a pick is still open: %+v", left)
	}
}

// A duplicate pick is refused, and refusing it changes nothing.
func TestDifferentColorsPromptRefusesARepeatedColor(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	vessel := pushDifferentSource(g, me)
	if err := g.ActivateManaAbility(me.ID, vessel, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if err := g.ResolveManaChoice(openManaPicks(g)[0].ID, me.ID, "R"); err != nil {
		t.Fatalf("answer R: %v", err)
	}
	second := openManaPicks(g)[0]
	if err := g.ResolveManaChoice(second.ID, me.ID, "R"); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("a second R = %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("a refused answer added mana: %v", me.ManaPool)
	}
	if again := openManaPicks(g); len(again) != 1 || again[0].ID != second.ID {
		t.Errorf("the refused answer moved the question: %+v", again)
	}
	// Colourless was never on offer.
	if err := g.ResolveManaChoice(second.ID, me.ID, "C"); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("C = %v, want ErrInvalidParam", err)
	}
}

// Named up front (#1443), the two colours must differ; a repeat is
// refused before the source taps.
func TestDifferentColorsUpfrontRefusesARepeatAndTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	vessel := pushDifferentSource(g, me)

	var lists [][]string
	g.WithWriteLock(func() {
		lists = ManaAbilityColorOptions(g, me.ID, vessel, differentShape()[0])
	})
	if len(lists) != 2 || len(lists[0]) != 5 || len(lists[1]) != 5 {
		t.Fatalf("color options = %v, want two lists of five, one per mana", lists)
	}
	err := g.ActivateManaAbility(me.ID, vessel, 0, ManaAbilityParams{Colors: []string{"B", "B"}})
	if !errors.Is(err, ErrIllegalManaColor) {
		t.Fatalf("[B B] = %v, want ErrIllegalManaColor", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == vessel && c.Tapped {
			t.Fatal("a refused colour pair tapped the source")
		}
	}
	if err := g.ActivateManaAbility(me.ID, vessel, 0, ManaAbilityParams{Colors: []string{"B", "W"}}); err != nil {
		t.Fatalf("[B W]: %v", err)
	}
	if got := sortedPool(me); !reflect.DeepEqual(got, []string{"B", "W"}) {
		t.Errorf("pool = %v, want [B W]", got)
	}
	if picks := openManaPicks(g); len(picks) != 0 {
		t.Errorf("colours named up front still queued a pick: %+v", picks)
	}
}

// The auto-tapper books a different pair that pays the cost, and the
// executor mints exactly that pair.
func TestAutoTapDifferentColorsPaysTwoColoursFromOneSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	vessel := pushDifferentSource(g, me)

	cost, err := ParseCost("{B}{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != vessel {
		t.Fatalf("plan = %v, %v; want the Vessel alone", plan, ok)
	}
	g.WithWriteLock(func() {
		full, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("autoTapLocked disagreed with AutoTapForCost")
		}
		if !reflect.DeepEqual(full[0].DifferentColors, []string{"B", "G"}) {
			t.Errorf("booked %v, want [B G]", full[0].DifferentColors)
		}
		g.materializePlanLocked(me, full, cost)
	})
	if got := sortedPool(me); !reflect.DeepEqual(got, []string{"B", "G"}) {
		t.Errorf("pool = %v, want [B G]", got)
	}
	if picks := openManaPicks(g); len(picks) != 0 {
		t.Errorf("the auto-tapper left a pick open: %+v", picks)
	}
}

// Two mana of one colour is not two mana of different colors: a lone
// Vessel never pays {U}{U}, and with an Island beside it pays {U}{U}{R}
// by booking blue and red.
func TestAutoTapDifferentColorsNeverPaysTwoOfOneColour(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushDifferentSource(g, me)

	same, err := ParseCost("{U}{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, same, 0); ok {
		t.Fatalf("plan = %v for {U}{U} off one Vessel — its colours differ", plan)
	}
	generic, err := ParseCost("{2}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if _, ok := g.AutoTapForCost(me.ID, generic, 0); !ok {
		t.Error("no plan for {2} off one Vessel")
	}

	island := pushIslandFor(g, me)
	cost, err := ParseCost("{U}{U}{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 2 {
		t.Fatalf("plan = %v, %v; want the Vessel and the Island", plan, ok)
	}
	g.WithWriteLock(func() {
		full, _ := g.autoTapLocked(me.ID, cost, 0, nil)
		g.materializePlanLocked(me, full, cost)
	})
	if !me.ManaPool.CanPayFor(cost, 0, ManaSpendContext{}) {
		t.Errorf("pool %v cannot pay {U}{U}{R}", me.ManaPool)
	}
	_ = island
}

// pickDifferentColors pays the most restrictive requirement first and
// never repeats a colour, even when two requirements ask for one.
func TestPickDifferentColorsNeverRepeats(t *testing.T) {
	cost, err := ParseCost("{U}{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	pending := append([]ColorRequirement(nil), cost.Required...)
	got := pickDifferentColors(AllColors, 2, &pending)
	if len(got) != 2 || got[0] == got[1] || got[0] != "U" {
		t.Errorf("picked %v, want blue and one other colour", got)
	}
	if len(pending) != 1 {
		t.Errorf("pending = %v, want one {U} still owed", pending)
	}
	if sets := differentColorSets(AllColors, 2); len(sets) != 10 {
		t.Errorf("%d pairs of five colours, want 10", len(sets))
	}
}
