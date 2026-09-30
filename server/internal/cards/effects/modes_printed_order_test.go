package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_printed_order_test.go — #1653, CR 608.2c: the chosen modes of
// a spell, a triggered ability and an activated ability are carried
// out in the order they are WRITTEN, whatever order the player
// announced them in, and each occurrence still reads its own target
// group. The fixtures log what each bullet did and to whom.

// printedOrderLog collects "<bullet>:<target>" lines as bullets run.
var printedOrderLog []string

func logBullet(bullet string, targets map[uuid.UUID]string) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok {
			printedOrderLog = append(printedOrderLog, bullet+":none")
			return nil
		}
		printedOrderLog = append(printedOrderLog, bullet+":"+targets[t.ID])
		return nil
	}
}

// printedOrderModes is "A (target creature you control) / B (target
// creature an opponent controls)"; the names come from `names`.
func printedOrderModes(names map[uuid.UUID]string, repeatable bool, n int) *game.ModeSpec {
	a := ModeDoing("Bullet A.", TargetCreature("target creature you control", YouControl()), logBullet("A", names))
	b := ModeDoing("Bullet B.", TargetCreature("target creature an opponent controls", OpponentControls()), logBullet("B", names))
	if repeatable {
		return ChooseNRepeating("Choose "+string(rune('0'+n)), n, n, a, b)
	}
	return ChooseN("Choose two", n, n, a, b)
}

const (
	printedOrderSpellOracle     = "test-printed-order-spell"
	printedOrderRepeatOracle    = "test-printed-order-repeat"
	printedOrderActivatedOracle = "test-printed-order-activated"
	printedOrderTriggerOracle   = "test-printed-order-trigger"
)

func TestPrintedOrderSpellRunsReversedAnnounceInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	names := map[uuid.UUID]string{mine: "mine", theirs: "theirs"}
	registerForTest(t, Spec{
		OracleID: printedOrderSpellOracle,
		Name:     "Printed Order Spell",
		Modes:    printedOrderModes(names, false, 2),
	})
	id := handCardFull(me, "Printed Order Spell", "Instant", "", printedOrderSpellOracle, nil)
	printedOrderLog = nil
	// Announced B then A: occurrence 0 is B (their creature), 1 is A.
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes: []int{1, 0},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: theirs, Mode: 0},
			{Kind: game.TargetCard, ID: mine, Mode: 1},
		},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	want := []string{"A:mine", "B:theirs"}
	if !slices.Equal(printedOrderLog, want) {
		t.Errorf("ran %v, want %v", printedOrderLog, want)
	}
}

// A repeatable mode: the option's occurrences keep their announce
// order relative to each other (CR 700.2d), each with its own target.
func TestPrintedOrderRepeatableModeKeepsRelativeAnnounceOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	t1 := pushVanillaCreature(g, opp.ID, "Theirs1", 1, 1)
	t2 := pushVanillaCreature(g, opp.ID, "Theirs2", 1, 1)
	names := map[uuid.UUID]string{mine: "mine", t1: "theirs1", t2: "theirs2"}
	registerForTest(t, Spec{
		OracleID: printedOrderRepeatOracle,
		Name:     "Printed Order Repeat",
		Modes:    printedOrderModes(names, true, 3),
	})
	id := handCardFull(me, "Printed Order Repeat", "Instant", "", printedOrderRepeatOracle, nil)
	printedOrderLog = nil
	// B (theirs2), A (mine), B (theirs1).
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes: []int{1, 0, 1},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: t2, Mode: 0},
			{Kind: game.TargetCard, ID: mine, Mode: 1},
			{Kind: game.TargetCard, ID: t1, Mode: 2},
		},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	want := []string{"A:mine", "B:theirs2", "B:theirs1"}
	if !slices.Equal(printedOrderLog, want) {
		t.Errorf("ran %v, want %v", printedOrderLog, want)
	}
}

func TestPrintedOrderActivatedAbilityRunsInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	names := map[uuid.UUID]string{mine: "mine", theirs: "theirs"}
	registerForTest(t, Spec{
		OracleID: printedOrderActivatedOracle,
		Name:     "Printed Order Activated",
		Activated: []ActivatedAbility{{
			Label: "Pay 1 life: Choose two",
			Cost:  PayLife(1),
			Modes: printedOrderModes(names, false, 2),
		}},
	})
	advanceToMain(t, g)
	src := pushCatalogPermanent(g, me.ID, "Printed Order Activated", "Artifact", printedOrderActivatedOracle, false)
	printedOrderLog = nil
	err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{
		Modes: []int{1, 0},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: theirs, Mode: 0},
			{Kind: game.TargetCard, ID: mine, Mode: 1},
		},
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	want := []string{"A:mine", "B:theirs"}
	if !slices.Equal(printedOrderLog, want) {
		t.Errorf("ran %v, want %v", printedOrderLog, want)
	}
}

func TestPrintedOrderTriggeredAbilityRunsInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	names := map[uuid.UUID]string{mine: "mine", theirs: "theirs"}
	tr := WhenThisEnters("Printed Order Trigger — choose two", func(*game.Game, *game.StackItem) error { return nil })
	tr.Modes = printedOrderModes(names, false, 2)
	registerForTest(t, Spec{
		OracleID:  printedOrderTriggerOracle,
		Name:      "Printed Order Trigger",
		Triggered: []game.TriggeredAbility{tr},
	})
	printedOrderLog = nil
	castCatalogSpell(t, g, "Printed Order Trigger", "Creature — Test", printedOrderTriggerOracle, nil)
	passPriorityAroundTable(t, g)
	p := pendingOfKind(g, game.PendingChoiceModePick)
	if p == nil {
		t.Fatal("the trigger asks for its modes")
	}
	if err := g.ResolveModePick(p.ID, me.ID, []int{1, 0}); err != nil {
		t.Fatalf("answer B then A: %v", err)
	}
	// One pick_target prompt per occurrence; answer each with the only
	// creature its clause offers.
	for i := 0; i < 4; i++ {
		pick := latestPickTarget(g, me.ID)
		if pick == nil {
			break
		}
		target := mine
		if slices.Contains(pick.PickTargetCards, theirs) {
			target = theirs
		}
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	want := []string{"A:mine", "B:theirs"}
	if !slices.Equal(printedOrderLog, want) {
		t.Errorf("ran %v, want %v", printedOrderLog, want)
	}
}

func TestPrintedModeOrder(t *testing.T) {
	got := game.PrintedModeOrder([]int{2, 0, 2, 1, 0})
	want := []int{1, 4, 3, 0, 2}
	if !slices.Equal(got, want) {
		t.Errorf("PrintedModeOrder = %v, want %v", got, want)
	}
}
