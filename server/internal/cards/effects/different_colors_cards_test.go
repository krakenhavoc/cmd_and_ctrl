package effects

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// different_colors_cards_test.go — #2558's cards: Firemind Vessel,
// Guild Globe, Interplanar Beacon's filter and Component Pouch, each
// "Add two mana of different colors". The engine half (the prompt, the
// up-front check, the auto-tapper) is tested in
// game/mana_different_colors_test.go; this file is the catalog side,
// the wire and the restore.

const (
	dcFiremindVesselOracle    = "d8afd7b0-7cd0-4caf-ac43-a0e24a93ca09"
	dcGuildGlobeOracle        = "a7e7cacf-cc2f-45ce-abb1-7b66d1e2fabc"
	dcInterplanarBeaconOracle = "073169f2-da3a-4a93-8c01-b3fd8558d225"
	dcComponentPouchOracle    = "4bb1863e-50f3-4e08-884e-1f58ee55817f"
)

func dcSortedPool(p *game.Player) []string {
	out := make([]string, 0, len(p.ManaPool))
	for _, tok := range p.ManaPool {
		out = append(out, tok.Color)
	}
	sort.Strings(out)
	return out
}

func dcManaPicks(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana {
			out = append(out, c)
		}
	}
	return out
}

// Firemind Vessel enters tapped, and taps for two mana of different
// colours: asked twice through the prompt, the second time without the
// first colour.
func TestFiremindVesselEntersTappedAndAddsTwoDifferentColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, "Firemind Vessel", "Artifact", dcFiremindVesselOracle, nil)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		t.Fatal("Firemind Vessel did not enter")
	}
	if !c.Tapped {
		t.Error("Firemind Vessel enters tapped")
	}
	c.Tapped = false

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := dcManaPicks(g)
	if len(picks) != 1 || picks[0].ManaDifferent != 2 {
		t.Fatalf("picks = %+v, want one pick of two different colours", picks)
	}
	if err := g.ResolveManaChoice(picks[0].ID, me.ID, "R"); err != nil {
		t.Fatalf("answer R: %v", err)
	}
	second := dcManaPicks(g)[0]
	if err := g.ResolveManaChoice(second.ID, me.ID, "R"); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("red twice = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveManaChoice(second.ID, me.ID, "W"); err != nil {
		t.Fatalf("answer W: %v", err)
	}
	if got := dcSortedPool(me); !reflect.DeepEqual(got, []string{"R", "W"}) {
		t.Errorf("pool = %v, want [R W]", got)
	}
}

// The wire publishes one colour list per mana and different_colors, so
// a client offers only pairs of different colours.
func TestFiremindVesselWireMarksDifferentColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Firemind Vessel", "Artifact", dcFiremindVesselOracle, false)
	row := upfrontManaRow(t, g, me.ID, id, 0)
	if !row.DifferentColors {
		t.Error("different_colors is not set")
	}
	if len(row.ColorOptions) != 2 || len(row.ColorOptions[0]) != 5 || len(row.ColorOptions[1]) != 5 {
		t.Errorf("color_options = %v, want two lists of five", row.ColorOptions)
	}
	// An ordinary any-colour source carries no such mark.
	birds := pushCatalogPermanent(g, me.ID, "Birds of Paradise", "Creature — Bird", upfrontBirdsOracle, false)
	if upfrontManaRow(t, g, me.ID, birds, 0).DifferentColors {
		t.Error("Birds of Paradise is marked different_colors")
	}
}

// A restore with the pick half answered resumes it. The colour already
// named stays named, and the rest of the answer adds both mana.
func TestFiremindVesselPickSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Firemind Vessel", "Artifact", dcFiremindVesselOracle, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if err := g.ResolveManaChoice(dcManaPicks(g)[0].ID, me.ID, "G"); err != nil {
		t.Fatalf("answer G: %v", err)
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not restorable: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back game.GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	picks := dcManaPicks(restored)
	if len(picks) != 1 {
		t.Fatalf("restored %d mana picks, want 1", len(picks))
	}
	pick := picks[0]
	if pick.ManaDifferent != 2 || !reflect.DeepEqual(pick.ManaChosen, []string{"G"}) {
		t.Fatalf("restored pick = %+v, want green already chosen", pick)
	}
	for _, c := range pick.ColorOptions {
		if c == "G" {
			t.Fatalf("restored pick offers green again: %v", pick.ColorOptions)
		}
	}
	rme := restored.Seats[0]
	if err := restored.ResolveManaChoice(pick.ID, rme.ID, "G"); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("green again after the restore = %v, want ErrInvalidParam", err)
	}
	if err := restored.ResolveManaChoice(pick.ID, rme.ID, "U"); err != nil {
		t.Fatalf("answer U after the restore: %v", err)
	}
	if got := dcSortedPool(rme); !reflect.DeepEqual(got, []string{"G", "U"}) {
		t.Errorf("pool = %v, want [G U]", got)
	}
}

// Guild Globe draws on entering, and its {2}, {T}, sacrifice filters
// two mana into two of different colours; the Globe is gone.
func TestGuildGlobeDrawsThenCracksForTwoDifferentColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, "Guild Globe", "Artifact", dcGuildGlobeOracle, nil)
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, id) == nil {
		t.Fatal("Guild Globe did not enter")
	}
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand = %d, want %d (one card drawn)", got, before+1)
	}

	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B", "R"}}); err == nil {
		t.Fatal("no {2} floating: the activation should be refused")
	}
	samiGiveColorless(me, 2)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U", "U"}}); !errors.Is(err, game.ErrIllegalManaColor) {
		t.Fatalf("blue twice = %v, want ErrIllegalManaColor", err)
	}
	if findBattlefieldCardForTest(g, id) == nil {
		t.Fatal("a refused pair sacrificed the Globe")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B", "R"}}); err != nil {
		t.Fatalf("ActivateManaAbility [B R]: %v", err)
	}
	if got := dcSortedPool(me); !reflect.DeepEqual(got, []string{"B", "R"}) {
		t.Errorf("pool = %v, want [B R]: the {2} paid, two different colours added", got)
	}
	if findBattlefieldCardForTest(g, id) != nil {
		t.Error("the Globe is sacrificed")
	}
}

// Interplanar Beacon's filter is back: {1}, {T} for two mana of
// different colours, each spendable only on a planeswalker spell.
func TestInterplanarBeaconFiltersIntoTwoRestrictedColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beacon := pushCatalogPermanent(g, me.ID, "Interplanar Beacon", "Land", dcInterplanarBeaconOracle, false)
	row := upfrontManaRow(t, g, me.ID, beacon, 1)
	if !row.DifferentColors {
		t.Fatal("the filter row is not marked different_colors")
	}
	samiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, beacon, 1, game.ManaAbilityParams{Colors: []string{"W", "U"}}); err != nil {
		t.Fatalf("ActivateManaAbility [W U]: %v", err)
	}
	if got := dcSortedPool(me); !reflect.DeepEqual(got, []string{"U", "W"}) {
		t.Fatalf("pool = %v, want [U W]", got)
	}
	want := []string{ManaRestrictCast, ManaRestrictType("Planeswalker")}
	for _, tok := range me.ManaPool {
		if !reflect.DeepEqual(tok.Restrictions, want) {
			t.Errorf("token %s restrictions = %v, want %v", tok.Color, tok.Restrictions, want)
		}
	}
}

// Interplanar Beacon's prompt path carries the restriction too, onto
// both mana, minted together.
func TestInterplanarBeaconPromptCarriesTheRestriction(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beacon := pushCatalogPermanent(g, me.ID, "Interplanar Beacon", "Land", dcInterplanarBeaconOracle, false)
	samiGiveColorless(me, 1)
	if err := g.ActivateManaAbility(me.ID, beacon, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if err := g.ResolveManaChoice(dcManaPicks(g)[0].ID, me.ID, "B"); err != nil {
		t.Fatalf("answer B: %v", err)
	}
	if err := g.ResolveManaChoice(dcManaPicks(g)[0].ID, me.ID, "G"); err != nil {
		t.Fatalf("answer G: %v", err)
	}
	if got := dcSortedPool(me); !reflect.DeepEqual(got, []string{"B", "G"}) {
		t.Fatalf("pool = %v, want [B G]", got)
	}
	for _, tok := range me.ManaPool {
		if len(tok.Restrictions) == 0 {
			t.Errorf("token %s lost the planeswalker-only restriction", tok.Color)
		}
	}
}

// Component Pouch: no counter, no mana; a counter buys two mana of
// different colours; the roll stocks one or two counters.
func TestComponentPouchSpendsACounterForTwoDifferentColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pouch := pushCatalogPermanent(g, me.ID, "Component Pouch", "Artifact", dcComponentPouchOracle, false)
	if err := g.ActivateManaAbility(me.ID, pouch, 0, game.ManaAbilityParams{Colors: []string{"W", "B"}}); err == nil {
		t.Fatal("no component counter: the mana ability should be refused")
	}
	if c := findBattlefieldCardForTest(g, pouch); c == nil || c.Tapped {
		t.Fatal("a refused activation tapped the pouch")
	}

	// The roll: 1–9 is one counter, 10–20 two.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, pouch, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("roll: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := counterOn(g, pouch, "component")
	if n != 1 && n != 2 {
		t.Fatalf("component counters after one roll = %d, want 1 or 2", n)
	}
	var roll int
	for _, ev := range g.Events {
		if ev.Kind == game.EventRollDie && ev.Source == pouch {
			roll = ev.Amount
		}
	}
	if roll != 0 && (roll >= 10) != (n == 2) {
		t.Errorf("rolled %d and got %d counters", roll, n)
	}

	c := findBattlefieldCardForTest(g, pouch)
	c.Tapped = false
	if err := g.ActivateManaAbility(me.ID, pouch, 0, game.ManaAbilityParams{Colors: []string{"W", "B"}}); err != nil {
		t.Fatalf("ActivateManaAbility [W B]: %v", err)
	}
	if got := dcSortedPool(me); !reflect.DeepEqual(got, []string{"B", "W"}) {
		t.Errorf("pool = %v, want [B W]", got)
	}
	if got := counterOn(g, pouch, "component"); got != n-1 {
		t.Errorf("component counters = %d, want %d", got, n-1)
	}
}

// The auto-tapper casts a two-colour spell off Firemind Vessel alone.
func TestFiremindVesselAutoTapsForATwoColourSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Firemind Vessel", "Artifact", dcFiremindVesselOracle, false)
	cost, err := game.ParseCost("{W}{B}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); !ok {
		t.Error("no plan for {W}{B} off Firemind Vessel")
	}
	same, _ := game.ParseCost("{B}{B}")
	if plan, ok := g.AutoTapForCost(me.ID, same, 0); ok {
		t.Errorf("plan %v for {B}{B} off Firemind Vessel — its colours differ", plan)
	}
	_ = uuid.Nil
}
