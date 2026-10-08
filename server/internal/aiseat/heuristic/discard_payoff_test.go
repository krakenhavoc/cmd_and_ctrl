package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// discard_payoff_test.go is ADR 0126's amendment of 2026-10-06: a
// discard that one of the bot's own triggered rows pays for costs what
// the payoff pays less.

// withDiscardPayoff gives a permanent a triggered row with that payoff.
func withDiscardPayoff(c protocol.CardView, d protocol.DiscardPayoffView) protocol.CardView {
	c.AbilityRows = append(c.AbilityRows, protocol.AbilityRowView{
		Kind: "triggered", Label: "discard payoff",
		Purpose: &protocol.PurposeView{DiscardPayoff: &d},
	})
	return c
}

// maryRead is Mary Read and Anne Bonny on the battlefield, with her
// Treasure for an Island, Pirate or Vehicle card.
func maryRead(id string, controller int) protocol.CardView {
	m := creature(id, controller, "Mary Read and Anne Bonny", 3, 3)
	m.TypeLine = "Legendary Creature — Human Assassin Pirate"
	return withDiscardPayoff(m, protocol.DiscardPayoffView{Types: []string{"island", "pirate", "vehicle"}, Tokens: 1})
}

func island(id string, controller int) protocol.CardView {
	c := land(id, controller)
	c.Name, c.TypeLine = "Island", "Basic Land — Island"
	return c
}

func noPayoffs() heuristic.Config {
	c := heuristic.DefaultConfig()
	c.PriceDiscardPayoffs = false
	return c
}

func windfallFixture(t *testing.T) (protocol.CardView, func(string) legal.Move) {
	windfall := spell(cardID(1), 0, "Unexpected Windfall", "{2}{R}{R}")
	windfall.Purpose = &protocol.PurposeView{Draws: 2, Tokens: 2}
	windfall.AdditionalCost = &protocol.AdditionalCostView{DiscardCards: 1}
	return windfall, func(discard string) legal.Move {
		m := castMove(t, 0, windfall.InstanceID, "Cast Unexpected Windfall discarding "+discard)
		m.Params = mustJSON(t, map[string]any{"instance_id": windfall.InstanceID, "from_zone": "hand", "discard_ids": []string{discard}})
		return m
	}
}

// The owner's position: with Mary Read out, Windfall pitches the
// Island, not the Mountain, because the Island makes a Treasure.
func TestADiscardPayoffMakesTheMatchingCardCheaper(t *testing.T) {
	windfall, cast := windfallFixture(t)
	mountain, isl := land(cardID(2), 0), island(cardID(3), 0)
	bf := append(manaLands(7, 0, 100), maryRead(cardID(4), 0))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, mountain, isl)), newSeat(1)},
		withBattlefield(bf...), withTurn(9, 0, "precombat_main"))
	in := input(0, v, passMove(0), cast(mountain.InstanceID), cast(isl.InstanceID))

	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Unexpected Windfall discarding "+isl.InstanceID {
		t.Fatalf("chose %q, want the Island discarded", got)
	}
	pol := heuristic.New()
	gap := rankValue(t, pol, in, in.Moves[2].Label) - rankValue(t, pol, in, in.Moves[1].Label)
	if want := heuristic.DefaultConfig().TokenWeight; !nearly(gap, want) {
		t.Errorf("the Island is %.3f cheaper than the Mountain, want one token's %.3f", gap, want)
	}
	off := heuristic.NewWithConfig(noPayoffs())
	if gap := rankValue(t, off, in, in.Moves[2].Label) - rankValue(t, off, in, in.Moves[1].Label); !nearly(gap, 0) {
		t.Errorf("with payoffs off the two lands differ by %.3f, want 0", gap)
	}
}

// What each kind of payoff is worth, and that only the bot's own
// permanents pay.
func TestADiscardPayoffIsPricedByWhatItPays(t *testing.T) {
	windfall, cast := windfallFixture(t)
	mountain := land(cardID(2), 0)
	w := heuristic.DefaultWeights()
	cfg := heuristic.DefaultConfig()
	price := func(pol *heuristic.Policy, extra ...protocol.CardView) float64 {
		bf := append(manaLands(7, 0, 100), extra...)
		v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, mountain)), newSeat(1), newSeat(2), newSeat(3)},
			withBattlefield(bf...), withTurn(9, 0, "precombat_main"))
		in := input(0, v, passMove(0), cast(mountain.InstanceID))
		return rankValue(t, pol, in, in.Moves[1].Label)
	}
	base := price(heuristic.New())
	mako := withDiscardPayoff(creature(cardID(5), 0, "Marauding Mako", 1, 1), protocol.DiscardPayoffView{Any: true, Counters: 1})
	glint := withDiscardPayoff(creature(cardID(6), 0, "Glint-Horn Buccaneer", 2, 4), protocol.DiscardPayoffView{Any: true, DamageEachOpponent: 1})
	cases := []struct {
		name string
		bf   protocol.CardView
		want float64
	}{
		{"Mako's counter", mako, w.Power + w.Toughness},
		{"Glint-Horn's damage to each of three opponents", glint, 3 * cfg.DamageToOpponent},
		{"Mary Read pays nothing for a Mountain", maryRead(cardID(7), 0), 0},
		{"an opponent's Mako pays the bot nothing", withDiscardPayoff(creature(cardID(8), 1, "Marauding Mako", 1, 1), protocol.DiscardPayoffView{Any: true, Counters: 1}), 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := price(heuristic.New(), c.bf) - base
			// The board value of the permanent itself does not enter a
			// cast's price, so the difference is the payoff alone.
			if !nearly(got, c.want) {
				t.Errorf("the payoff added %.3f, want %.3f", got, c.want)
			}
			if off := price(heuristic.NewWithConfig(noPayoffs()), c.bf) - price(heuristic.NewWithConfig(noPayoffs())); !nearly(off, 0) {
				t.Errorf("with payoffs off it added %.3f", off)
			}
		})
	}
}

// A loot's discard names the card the payoff pays for, and the loot
// itself is worth the payoff it will trigger.
func TestALootDiscardsWhatThePayoffPaysFor(t *testing.T) {
	const choiceID = "choice-loot"
	mountain, isl := land(cardID(2), 0), island(cardID(3), 0)
	mary := maryRead(cardID(4), 0)
	t.Run("the choice", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(isl, mountain)), newSeat(1)},
			withBattlefield(append(sixLands(), mary)...),
			withChoice(handChoice(choiceID, "Mary Read and Anne Bonny — discard a card", 1, 1, isl, mountain)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "pitch the mountain", map[string]any{"card_ids": []string{mountain.InstanceID}}),
			choiceMove(t, 0, choiceID, "pitch the island", map[string]any{"card_ids": []string{isl.InstanceID}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the island" {
			t.Fatalf("chose %q, want the Island, which makes a Treasure", got)
		}
	})
	t.Run("the activation", func(t *testing.T) {
		mary.ActivatedAbilities = []protocol.ActivatedAbilityView{{
			Index: 0, TapCost: true, Purpose: &protocol.PurposeView{Draws: 1, Discards: 1},
		}}
		loot := legal.Move{
			Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
			Label: "Loot", Source: uuid.MustParse(mary.InstanceID),
			Params: mustJSON(t, map[string]any{"source_card_id": mary.InstanceID, "ability_index": 0}),
		}
		at := func(pol *heuristic.Policy, hand ...protocol.CardView) float64 {
			v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)},
				withBattlefield(append(sixLands(), mary)...), withTurn(5, 0, "postcombat_main"))
			return rankValue(t, pol, input(0, v, passMove(0), loot), "Loot")
		}
		gap := at(heuristic.New(), isl, mountain) - at(heuristic.New(), mountain)
		if want := heuristic.DefaultConfig().TokenWeight; !nearly(gap, want) {
			t.Errorf("an Island in hand adds %.3f to the loot, want one token's %.3f", gap, want)
		}
		off := heuristic.NewWithConfig(noPayoffs())
		if gap := at(off, isl, mountain) - at(off, mountain); !nearly(gap, 0) {
			t.Errorf("with payoffs off an Island adds %.3f to the loot", gap)
		}
	})
}

// The cleanup-step discard to hand size pitches the payoff's card too.
func TestCleanupDiscardPitchesThePayoffsCard(t *testing.T) {
	mountain, isl := land(cardID(2), 0), island(cardID(3), 0)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(isl, mountain)), newSeat(1)},
		withBattlefield(append(sixLands(), maryRead(cardID(4), 0))...), withTurn(5, 0, "cleanup"))
	in := input(0, v, discardMove(t, 0, "discard the mountain", mountain.InstanceID), discardMove(t, 0, "discard the island", isl.InstanceID))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard the island" {
		t.Fatalf("chose %q, want the Island", got)
	}
}

// ADR 0135 §2: an alternative cost that DISCARDS (Foil's, Snag's) pays
// the discard payoffs as an additional cost's discard does, so with Mary
// Read out the bot pays with the Island. An offer that exiles its card
// (a pitch) pays none.
func TestADiscardAlternativeCostPaysTheDiscardPayoff(t *testing.T) {
	for _, tc := range []struct {
		name     string
		discards bool
		want     float64
	}{
		{"discard", true, heuristic.DefaultConfig().TokenWeight},
		{"pitch (an exile)", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sp := spell(cardID(1), 0, "Discard Counter", "{2}{U}{U}")
			sp.AlternativeCosts = []protocol.AlternativeCostView{{Key: "alt", Discards: tc.discards}}
			mountain, isl := land(cardID(2), 0), island(cardID(3), 0)
			cast := func(pay string) legal.Move {
				m := castMove(t, 0, sp.InstanceID, "Cast for its alternative cost paying "+pay)
				m.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand",
					"alternative_cost": "alt", "alt_cost_ids": []string{pay}})
				return m
			}
			bf := append(manaLands(7, 0, 100), maryRead(cardID(4), 0))
			v := newView([]protocol.PlayerView{newSeat(0, withHand(sp, mountain, isl)), newSeat(1)},
				withBattlefield(bf...), withTurn(9, 0, "precombat_main"))
			in := input(0, v, passMove(0), cast(mountain.InstanceID), cast(isl.InstanceID))
			pol := heuristic.New()
			gap := rankValue(t, pol, in, in.Moves[2].Label) - rankValue(t, pol, in, in.Moves[1].Label)
			if !nearly(gap, tc.want) {
				t.Errorf("the Island is %.3f cheaper than the Mountain, want %.3f", gap, tc.want)
			}
		})
	}
}
