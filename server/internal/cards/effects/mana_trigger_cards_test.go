package effects

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_trigger_cards_test.go — the catalog half of #763: the cards
// CR 605.1b's triggered mana abilities unblocked. The engine rules
// themselves are pinned in game/mana_trigger_test.go; these are the
// printed cards, end to end, through the real registry.

const (
	wildGrowthOracle    = "706ae742-1807-44b7-a4fa-f2e26f61519a"
	overgrowthOracle    = "e6ccebf4-f4f0-404f-a634-4d751ef9c8aa"
	utopiaSprawlOracle  = "00d8efa6-a2d9-4249-8da7-b45173675329"
	fertileGroundOracle = "cf14d4e5-5965-45ad-97f7-26facf2884b5"
	manaFlareOracle     = "97159138-c34b-416e-b079-5c952383a243"
)

// pushAuraOnLand seeds an Aura already attached to `host`, the shape
// the cast path leaves behind once the Aura resolves.
func pushAuraOnLand(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Aura", OracleID: oracle,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect %s: %v", name, err)
		}
	})
	return id
}

// tapForMana activates a permanent's first mana ability.
func tapForMana(t *testing.T, g *game.Game, player, card uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(player, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
}

// sortedPool is poolColors sorted, for an assertion that does not care
// which order the tokens arrived in.
func sortedPool(p *game.Player) []string {
	out := poolColors(p)
	sort.Strings(out)
	return out
}

// Wild Growth: the Forest taps for {G}{G} and neither mana waits on
// the stack.
func TestWildGrowthAddsAnAdditionalGreenWithoutTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	pushAuraOnLand(t, g, me.ID, "Wild Growth", wildGrowthOracle, forest)

	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	aura := pushAuraOnLand(t, g, me.ID, "Wild Growth", wildGrowthOracle, forest)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want just the Forest's {G}", got)
	}
}

// Overgrowth adds two.
func TestOvergrowthAddsTwoGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	pushAuraOnLand(t, g, me.ID, "Overgrowth", overgrowthOracle, forest)

	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 3 {
		t.Errorf("pool = %v, want three green", got)
	}
}

// Utopia Sprawl: the colour chosen as it enters is the colour the
// trigger adds, and before the choice it adds nothing.
func TestUtopiaSprawlAddsTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	sprawl := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Utopia Sprawl", TypeLine: "Enchantment — Aura",
		OracleID: utopiaSprawlOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sprawl, game.TargetRef{Kind: game.TargetCard, ID: forest}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	// No colour chosen yet: the Sprawl adds nothing, never "any
	// colour".
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("pool = %v before the colour is chosen, want just the Forest's {G}", got)
	}
	untap(g, forest)
	g.WithWriteLock(func() { me.ManaPool = nil })

	// Choose blue, the way the entry hook does.
	g.WithWriteLock(func() {
		g.QueueColorChoiceForEffect(me.ID, sprawl, "Utopia Sprawl", nil, game.ColorForMana)
	})
	answerColor(t, g, me.ID, "U")

	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool = %v, want the Forest's {G} and the chosen {U}", got)
	}
}

// Fertile Ground's "any color" is a prompt on a hand-clicked tap.
func TestFertileGroundPromptsForTheAdditionalColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	pushAuraOnLand(t, g, me.ID, "Fertile Ground", fertileGroundOracle, forest)

	tapForMana(t, g, me.ID, forest)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for Fertile Ground's 'any color'")
	}
	if len(pick.ColorOptions) != 5 {
		t.Errorf("options = %v, want all five colours", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "R" {
		t.Errorf("pool = %v, want {G} and the chosen {R}", got)
	}
	if p := pendingOfKind(g, game.PendingChoiceMana); p != nil {
		t.Errorf("a second mana pick is open (%+v): the trigger must not re-trigger", p)
	}
}

// Mana Flare fires for a land ANY player taps, and gives the mana to
// that player, in a type the land actually produced.
func TestManaFlareDoublesEveryPlayersLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	them := g.Seats[1]
	b31Push(g, me.ID, "Mana Flare", "Enchantment", manaFlareOracle, "", 0, 0)
	mine := b31Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	theirs := b31Push(g, them.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)

	tapForMana(t, g, me.ID, mine)
	if got := sortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("my pool = %v, want two red", got)
	}
	tapForMana(t, g, them.ID, theirs)
	if got := sortedPool(them); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("their pool = %v, want two black — 'a player', not 'you'", got)
	}
	// One produced type is not a choice, so nobody was prompted.
	if p := pendingOfKind(g, game.PendingChoiceMana); p != nil {
		t.Errorf("a mana pick is open (%+v): one produced type is one option", p)
	}
}

// Mana Flare does NOT fire on a non-land source, and does not fire on
// mana an effect added (CR 106.12a).
func TestManaFlareIgnoresRocksAndSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b31Push(g, me.ID, "Mana Flare", "Enchantment", manaFlareOracle, "", 0, 0)
	// Sol Ring: an artifact, not a land.
	ring := b31Push(g, me.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487", "", 0, 0)

	tapForMana(t, g, me.ID, ring)
	if got := sortedPool(me); len(got) != 2 {
		t.Errorf("pool = %v, want Sol Ring's two {C} and nothing extra — it is not a land", got)
	}
	g.WithWriteLock(func() {
		me.ManaPool = nil
		if err := g.AddManaForEffect(me.ID, uuid.New(), "{B}{B}{B}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	if got := sortedPool(me); len(got) != 3 {
		t.Errorf("pool = %v, want Dark Ritual's three and nothing extra (CR 106.12a)", got)
	}
}

// The auto-tapper fires it too, so "Auto-tap & cast" and clicking the
// land by hand put the same mana in the pool.
func TestWildGrowthFiresUnderTheAutoTapper(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	pushAuraOnLand(t, g, me.ID, "Wild Growth", wildGrowthOracle, forest)

	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != forest {
		t.Fatalf("plan = %v (ok=%v), want the Forest", plan, ok)
	}
	// Tapping it is what runs the trigger; the planner never counted
	// the extra mana (ADR 0074 §7), so the second {G} simply floats.
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 2 {
		t.Errorf("pool = %v, want two green", got)
	}
}
