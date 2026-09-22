package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// charged_mana_cost_view_test.go — #1190: an activated ability's row
// and a mana ability's row both carry ChargedManaCost, the printed
// ManaCost run through the CR 601.2f cost-modifier pass the engine
// actually charges (#1184's AbilityManaCostForEffect, and #1191's
// ManaAbilityManaCostForEffect twin for the CR 605 half). Filed from
// #1193: the view used to stamp only the printed string, so a
// discounted ability's menu row and its real price could disagree.

// installChargedCostModifier hooks game.CatalogCostModifiers for the
// length of the test — the protocol-package twin of the `game`
// package's own costModifierCatalog test helper (exhaust_readers_test.go),
// needed here because CostModifiersFor reads the catalog hook and this
// package cannot reach that test file's unexported helper.
func installChargedCostModifier(t *testing.T, oracle string, mods ...game.CostModifier) {
	t.Helper()
	prev := game.CatalogCostModifiers
	game.CatalogCostModifiers = func(key string) []game.CostModifier {
		if key == oracle {
			return mods
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCostModifiers = prev })
}

// anotherPermanentDiscount is Boom Scholar's own clause, without the
// exhaust predicate half — these fixtures are ordinary (non-exhaust)
// abilities, so only "another permanent you control" is under test.
func anotherPermanentDiscount(amount int) game.CostModifier {
	return game.CostModifier{
		Kind:        game.CostReduction,
		Activations: true,
		Label:       "Abilities of other permanents you control cost less to activate.",
		AppliesTo: func(q game.CostQuery) bool {
			return q.Ability != nil && q.Card.Controller == q.Source.Controller &&
				q.Card.InstanceID != q.Source.InstanceID
		},
		Amount: func(game.CostQuery) int { return amount },
	}
}

// mustChargedCost dereferences ChargedManaCost or fails the test — a
// nil pointer here means the row fell back to the printed cost, which
// every case in this file expects NOT to happen once a source is on
// the board with a mana component.
func mustChargedCost(t *testing.T, p *string) string {
	t.Helper()
	if p == nil {
		t.Fatal("ChargedManaCost is nil, want a priced answer")
	}
	return *p
}

func TestActivatedAbilityRowShowsTheChargedCostAndThePrintedOneWhenTheyDiffer(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID

	target := game.NewCard("Discount Target", me)
	target.TypeLine = "Artifact"
	target.ActivatedAbilities = []game.ActivatedAbilityShape{{
		Label: "{3}{R}: Do a thing.",
		Cost:  game.AbilityCost{Mana: "{3}{R}"},
	}}
	g.Battlefield.PushTop(target)

	// Undiscounted: ChargedManaCost equals the printed ManaCost.
	row := battlefieldCardView(t, ViewOfGame(g), target.InstanceID).ActivatedAbilities[0]
	if row.ManaCost != "{3}{R}" {
		t.Fatalf("printed ManaCost = %q, want {3}{R}", row.ManaCost)
	}
	if got := mustChargedCost(t, row.ChargedManaCost); got != row.ManaCost {
		t.Errorf("with no discounter on the board, ChargedManaCost = %q, want it to equal the printed cost %q",
			got, row.ManaCost)
	}

	discounter := game.NewCard("Discounter", me)
	discounter.TypeLine = "Artifact"
	discounter.OracleID = "protocol-test-charged-cost-discounter"
	installChargedCostModifier(t, discounter.OracleID, anotherPermanentDiscount(2))
	g.Battlefield.PushTop(discounter)

	row = battlefieldCardView(t, ViewOfGame(g), target.InstanceID).ActivatedAbilities[0]
	if row.ManaCost != "{3}{R}" {
		t.Errorf("the discount must not touch the printed field: ManaCost = %q, want {3}{R}", row.ManaCost)
	}
	got := mustChargedCost(t, row.ChargedManaCost)
	if got != "{1}{R}" {
		t.Errorf("ChargedManaCost = %q, want {1}{R} — {3}{R} reduced by {2} generic", got)
	}
	if got == row.ManaCost {
		t.Errorf("ChargedManaCost and ManaCost read the same (%q) — the discount did not reach the row", row.ManaCost)
	}
}

func TestManaAbilityRowShowsTheChargedCostAndThePrintedOneWhenTheyDiffer(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID

	target := game.NewCard("Discount Mana Target", me)
	target.TypeLine = "Artifact"
	target.ManaAbilities = []game.ManaAbilityShape{{
		TapCost:  true,
		ManaCost: "{3}",
		Produced: "{C}{C}{C}",
		Label:    "{3}, {T}: Add {C}{C}{C}.",
	}}
	g.Battlefield.PushTop(target)

	row := battlefieldCardView(t, ViewOfGame(g), target.InstanceID).ManaAbilities[0]
	if row.ManaCost != "{3}" {
		t.Fatalf("printed ManaCost = %q, want {3}", row.ManaCost)
	}
	if got := mustChargedCost(t, row.ChargedManaCost); got != row.ManaCost {
		t.Errorf("with no discounter on the board, ChargedManaCost = %q, want it to equal the printed cost %q",
			got, row.ManaCost)
	}

	discounter := game.NewCard("Mana Discounter", me)
	discounter.TypeLine = "Artifact"
	discounter.OracleID = "protocol-test-charged-mana-cost-discounter"
	installChargedCostModifier(t, discounter.OracleID, anotherPermanentDiscount(2))
	g.Battlefield.PushTop(discounter)

	row = battlefieldCardView(t, ViewOfGame(g), target.InstanceID).ManaAbilities[0]
	if row.ManaCost != "{3}" {
		t.Errorf("the discount must not touch the printed field: ManaCost = %q, want {3}", row.ManaCost)
	}
	got := mustChargedCost(t, row.ChargedManaCost)
	if got != "{1}" {
		t.Errorf("ChargedManaCost = %q, want {1} — {3} reduced by {2} generic", got)
	}
	if got == row.ManaCost {
		t.Errorf("ChargedManaCost and ManaCost read the same (%q) — the discount did not reach the row", row.ManaCost)
	}
}

// TestAFullyDiscountedAbilityRowStampsARealEmptyStringNotNil is the
// property that made ChargedManaCost a pointer in the first place
// (mirroring LoyaltyCost, ActivatedAbilityView's own precedent): a
// discount that empties a mana component out completely is a REAL
// answer — "this now costs nothing" — and must not read the same on
// the wire as "the server could not price this ability at all".
func TestAFullyDiscountedAbilityRowStampsARealEmptyStringNotNil(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID

	target := game.NewCard("Free Now Target", me)
	target.TypeLine = "Artifact"
	target.ActivatedAbilities = []game.ActivatedAbilityShape{{
		Label: "{2}: Do a thing.",
		Cost:  game.AbilityCost{Mana: "{2}"},
	}}
	g.Battlefield.PushTop(target)

	discounter := game.NewCard("Free Now Discounter", me)
	discounter.TypeLine = "Artifact"
	discounter.OracleID = "protocol-test-fully-discounted-cost-discounter"
	installChargedCostModifier(t, discounter.OracleID, anotherPermanentDiscount(2))
	g.Battlefield.PushTop(discounter)

	row := battlefieldCardView(t, ViewOfGame(g), target.InstanceID).ActivatedAbilities[0]
	if row.ChargedManaCost == nil {
		t.Fatal("ChargedManaCost is nil, want a pointer to \"\" — the printed {2} was priced, not unpriceable")
	}
	if *row.ChargedManaCost != "" {
		t.Errorf("ChargedManaCost = %q, want \"\" — {2} reduced by {2} generic floors at nothing to render", *row.ChargedManaCost)
	}
}
