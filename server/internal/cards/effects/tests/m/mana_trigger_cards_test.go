package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Growth: the Forest taps for {G}{G} and neither mana waits on
// the stack.
func TestWildGrowthAddsAnAdditionalGreenWithoutTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.PushAuraOnLand(t, g, me.ID, "Wild Growth", effectstest.WildGrowthOracle, forest)

	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
	g.ReadSnapshot(func() {
		if len(g.PendingTriggers) != 0 {
			t.Errorf("PendingTriggers = %d: a triggered mana ability does not use the stack (CR 605.4a)",
				len(g.PendingTriggers))
		}
		if len(g.PendingChoices) != 0 {
			t.Errorf("PendingChoices = %d, want none", len(g.PendingChoices))
		}
	})
}

// Destroying the Aura stops the trigger with no bookkeeping, because
// the attachment is re-read on every production.
func TestWildGrowthStopsWhenTheAuraLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	aura := effectstest.PushAuraOnLand(t, g, me.ID, "Wild Growth", effectstest.WildGrowthOracle, forest)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want just the Forest's {G}", got)
	}
}

// Overgrowth adds two.
func TestOvergrowthAddsTwoGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.PushAuraOnLand(t, g, me.ID, "Overgrowth", effectstest.OvergrowthOracle, forest)

	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 3 {
		t.Errorf("pool = %v, want three green", got)
	}
}

// Utopia Sprawl: the colour chosen as it enters is the colour the
// trigger adds, and before the choice it adds nothing.
func TestUtopiaSprawlAddsTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	sprawl := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Utopia Sprawl", TypeLine: "Enchantment — Aura",
		OracleID: effectstest.UtopiaSprawlOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sprawl, game.TargetRef{Kind: game.TargetCard, ID: forest}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	// No colour chosen yet: the Sprawl adds nothing, never "any
	// colour".
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("pool = %v before the colour is chosen, want just the Forest's {G}", got)
	}
	effectstest.Untap(g, forest)
	g.WithWriteLock(func() { me.ManaPool = nil })

	// Choose blue, the way the entry hook does.
	g.WithWriteLock(func() {
		g.QueueColorChoiceForEffect(me.ID, sprawl, "Utopia Sprawl", nil, game.ColorForMana)
	})
	effectstest.AnswerColor(t, g, me.ID, "U")

	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool = %v, want the Forest's {G} and the chosen {U}", got)
	}
}

// Fertile Ground's "any color" is a prompt on a hand-clicked tap.
func TestFertileGroundPromptsForTheAdditionalColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.PushAuraOnLand(t, g, me.ID, "Fertile Ground", effectstest.FertileGroundOracle, forest)

	effectstest.TapForMana(t, g, me.ID, forest)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for Fertile Ground's 'any color'")
	}
	if len(pick.ColorOptions) != 5 {
		t.Errorf("options = %v, want all five colours", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "R" {
		t.Errorf("pool = %v, want {G} and the chosen {R}", got)
	}
	if p := effectstest.PendingOfKind(g, game.PendingChoiceMana); p != nil {
		t.Errorf("a second mana pick is open (%+v): the trigger must not re-trigger", p)
	}
}

// Mana Flare fires for a land ANY player taps, and gives the mana to
// that player, in a type the land actually produced.
func TestManaFlareDoublesEveryPlayersLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	them := g.Seats[1]
	effectstest.B31Push(g, me.ID, "Mana Flare", "Enchantment", effectstest.ManaFlareOracle, "", 0, 0)
	mine := effectstest.B31Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	theirs := effectstest.B31Push(g, them.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)

	effectstest.TapForMana(t, g, me.ID, mine)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("my pool = %v, want two red", got)
	}
	effectstest.TapForMana(t, g, them.ID, theirs)
	if got := effectstest.SortedPool(them); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("their pool = %v, want two black — 'a player', not 'you'", got)
	}
	// One produced type is not a choice, so nobody was prompted.
	if p := effectstest.PendingOfKind(g, game.PendingChoiceMana); p != nil {
		t.Errorf("a mana pick is open (%+v): one produced type is one option", p)
	}
}

// Mana Flare does NOT fire on a non-land source, and does not fire on
// mana an effect added (CR 106.12a).
func TestManaFlareIgnoresRocksAndSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B31Push(g, me.ID, "Mana Flare", "Enchantment", effectstest.ManaFlareOracle, "", 0, 0)
	// Sol Ring: an artifact, not a land.
	ring := effectstest.B31Push(g, me.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487", "", 0, 0)

	effectstest.TapForMana(t, g, me.ID, ring)
	if got := effectstest.SortedPool(me); len(got) != 2 {
		t.Errorf("pool = %v, want Sol Ring's two {C} and nothing extra — it is not a land", got)
	}
	g.WithWriteLock(func() {
		me.ManaPool = nil
		if err := g.AddManaForEffect(me.ID, uuid.New(), "{B}{B}{B}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	if got := effectstest.SortedPool(me); len(got) != 3 {
		t.Errorf("pool = %v, want Dark Ritual's three and nothing extra (CR 106.12a)", got)
	}
}

// The auto-tapper fires it too, so "Auto-tap & cast" and clicking the
// land by hand put the same mana in the pool.
func TestWildGrowthFiresUnderTheAutoTapper(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.PushAuraOnLand(t, g, me.ID, "Wild Growth", effectstest.WildGrowthOracle, forest)

	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != forest {
		t.Fatalf("plan = %v (ok=%v), want the Forest", plan, ok)
	}
	// Tapping it is what runs the trigger. The planner counts the extra
	// {G} too (ADR 0074, amendment of 2026-10-30), but {G} needs one
	// mana and a land cannot be half-tapped, so the second {G} simply
	// floats.
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 {
		t.Errorf("pool = %v, want two green", got)
	}
}
