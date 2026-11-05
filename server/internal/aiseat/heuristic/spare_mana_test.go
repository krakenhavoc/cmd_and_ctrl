package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spare_mana_test.go is #5085: Mind Stone's "{1}, {T}, Sacrifice Mind
// Stone: Draw a card" is cashed in once the bot has mana to spare and
// kept while its hand still wants the stone's mana. With
// PriceSpareManaSources off (the arena's `heuristic-nosparemana`) the
// stone costs its board value (1.00) against the card's 1.20, under
// PassThreshold, and the bot never cracks it, as in #4991's arena.

// mindStone is Mind Stone with its draw row as the view projects it.
func mindStone(id string) protocol.CardView {
	c := rock(id, 0, "Mind Stone", "{2}", protocol.ManaAbilityView{Index: 0, TapCost: true, Produced: "{C}"})
	value := []string{"value"}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "{1}, {T}, Sacrifice Mind Stone: Draw a card.", ManaCost: "{1}",
		TapCost: true, SacrificeSelf: true,
		Purpose: &protocol.PurposeView{Draws: 1, Answers: &value},
	}}
	return c
}

// stoneBoard is Mind Stone and n lands in the bot's own step, with the
// hand given, and the stone's draw on offer.
func stoneBoard(t *testing.T, step string, lands int, hand ...protocol.CardView) aiseat.Input {
	t.Helper()
	bf := append(manaLands(lands, 0, 100), mindStone(cardID(1)))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)}, withBattlefield(bf...), withTurn(9, 0, step))
	return input(0, v, passMove(0), activate(t, 0, cardID(1), "Mind Stone: draw a card", nil))
}

// noSpareMana is today's heuristic with the price off, as the arena's
// `heuristic-nosparemana` seats it.
func noSpareMana() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceSpareManaSources = false
	return heuristic.NewWithConfig(cfg)
}

// The issue's position: late, seven lands and an empty hand. The stone
// is spare, costs 0.50, and the card (1.20) clears the bar. Backed out,
// it costs 1.00 and the bot passes, as it did ~6,000 times in #4991.
func TestMindStoneCashedInWithManaToSpare(t *testing.T) {
	for _, step := range []string{"precombat_main", "postcombat_main"} {
		in := stoneBoard(t, step, 7)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Mind Stone: draw a card" {
			t.Errorf("%s: the heuristic chose %q, want the stone cashed in for a card", step, got)
		}
		if got := rankValue(t, heuristic.New(), in, "Mind Stone: draw a card"); !nearly(got, 0.70) {
			t.Errorf("%s: the draw priced %.3f, want 0.70 (a card less a spare stone)", step, got)
		}
		if got := chose(t, in, decide(t, noSpareMana(), in)); got != "Pass priority" {
			t.Errorf("%s: with the price off the heuristic chose %q, want the pass it made before #5085", step, got)
		}
	}
}

// Early, the stone's mana is what the hand needs: a five-drop with three
// lands, or nothing to cast but too few sources to give one up
// (SpareManaKeep). The bot keeps the stone.
func TestMindStoneKeptEarly(t *testing.T) {
	five := rampFiveDrop(cardID(50), 0)
	for name, in := range map[string]aiseat.Input{
		"a five-drop in hand, three lands": stoneBoard(t, "precombat_main", 3, five),
		"an empty hand, three lands":       stoneBoard(t, "precombat_main", 3),
		"an empty hand, four lands":        stoneBoard(t, "postcombat_main", 4),
	} {
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("%s: the heuristic chose %q, want the stone kept", name, got)
		}
	}
}

// The hand's want counts: with six lands a seven-drop still wants the
// stone's mana, and with seven lands it does not.
func TestMindStoneKeptWhileTheHandWantsItsMana(t *testing.T) {
	seven := creature(cardID(50), 0, "Seven Drop", 7, 7)
	seven.ManaCost = "{5}{G}{G}"
	if in := stoneBoard(t, "postcombat_main", 6, seven); chose(t, in, decide(t, heuristic.New(), in)) != "Pass priority" {
		t.Error("six lands and a seven-drop in hand: the stone was cracked, want it kept for the seventh mana")
	}
	if in := stoneBoard(t, "postcombat_main", 7, seven); chose(t, in, decide(t, heuristic.New(), in)) != "Mind Stone: draw a card" {
		t.Error("seven lands and a seven-drop in hand: the stone was kept, want it cashed in")
	}
}

// The price is the sacrifice's, not Mind Stone's: a rock another row
// sacrifices is spare by the same rule, and a mana creature never is.
func TestSpareRockForAnotherRowsSacrifice(t *testing.T) {
	outlet := func(id string) protocol.CardView {
		c := rock(id, 0, "Outlet", "{3}")
		c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
			Index: 0, Label: "{1}, Sacrifice an artifact or creature: Draw a card.", ManaCost: "{1}",
			SacrificeLabel: "an artifact or creature", Purpose: &protocol.PurposeView{Draws: 1},
		}}
		return c
	}
	board := func(lands int, fodder protocol.CardView) aiseat.Input {
		bf := append(manaLands(lands, 0, 100), outlet(cardID(1)), fodder)
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(bf...), withTurn(9, 0, "postcombat_main"))
		return input(0, v, passMove(0), sacMove(t, 0, cardID(1), "Outlet: draw a card", fodder.InstanceID))
	}
	signet := arcaneSignet(cardID(2), 0)
	if in := board(7, signet); chose(t, in, decide(t, heuristic.New(), in)) != "Outlet: draw a card" {
		t.Error("seven lands: the spare Signet was kept, want it sacrificed for a card")
	}
	if in := board(7, signet); chose(t, in, decide(t, noSpareMana(), in)) != "Pass priority" {
		t.Error("with the price off the spare Signet was sacrificed, want a pass")
	}
	if in := board(3, signet); chose(t, in, decide(t, heuristic.New(), in)) != "Pass priority" {
		t.Error("three lands: the Signet was sacrificed, want it kept")
	}
	if in := board(7, llanowarElves(cardID(2), 0)); chose(t, in, decide(t, heuristic.New(), in)) != "Pass priority" {
		t.Error("the Elves were sacrificed as a spare rock, want a mana creature priced by its body")
	}
}

// The switch is on by default, keeping five mana, and off in the baseline.
func TestPriceSpareManaSourcesDefaults(t *testing.T) {
	if cfg := heuristic.DefaultConfig(); !cfg.PriceSpareManaSources || cfg.SpareManaKeep != 5 || cfg.SpareManaSourceCost != 0.50 {
		t.Errorf("DefaultConfig: PriceSpareManaSources %v, SpareManaKeep %d, SpareManaSourceCost %.2f; want on, 5, 0.50",
			cfg.PriceSpareManaSources, cfg.SpareManaKeep, cfg.SpareManaSourceCost)
	}
	if cfg := heuristic.BaselineConfig(); cfg.PriceSpareManaSources || cfg.SpareManaKeep != 0 || cfg.SpareManaSourceCost != 0 {
		t.Errorf("BaselineConfig: PriceSpareManaSources %v, SpareManaKeep %d, SpareManaSourceCost %.2f; want all zero",
			cfg.PriceSpareManaSources, cfg.SpareManaKeep, cfg.SpareManaSourceCost)
	}
}
