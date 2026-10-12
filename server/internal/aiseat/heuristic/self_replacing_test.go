package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// self_replacing_test.go is ADR 0126's amendment of 2026-10-12 (#2457):
// a card whose hand row turns it into a card (cycling) is a cheaper
// discard, by the mana a discard spares cycling it, and never cheaper
// than the card cycling would find.

// withHandRow gives a card in hand a zone_abilities row that discards
// the card itself for `cost` and declares `purpose`.
func withHandRow(c protocol.CardView, cost string, purpose *protocol.PurposeView) protocol.CardView {
	c.ZoneAbilities = append(c.ZoneAbilities, protocol.ActivatedAbilityView{
		Index: len(c.ZoneAbilities), Label: "Cycling " + cost, ManaCost: cost, DiscardSelf: true, Purpose: purpose,
	})
	return c
}

// cycler is a card with "Cycling {2}", as the server projects it.
func cycler(c protocol.CardView) protocol.CardView {
	return withHandRow(c, "{2}", &protocol.PurposeView{Draws: 1})
}

func noSelfReplacing() heuristic.Config {
	c := heuristic.DefaultConfig()
	c.SelfReplacingDiscards = false
	return c
}

// The owner's position, reduced: Unexpected Windfall can pitch one of
// two 1/1s, one of them with cycling. Without the switch they cost the
// same. With it the cycler costs one card, the card cycling it would
// find, because a 1/1's body is worth less than the {2} cycling spares.
func TestACyclingCardIsACheaperDiscard(t *testing.T) {
	windfall, cast := windfallFixture(t)
	pilferer := creature(cardID(2), 0, "Impulsive Pilferer", 1, 1)
	mako := cycler(creature(cardID(3), 0, "Marauding Mako", 1, 1))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, pilferer, mako)), newSeat(1)},
		withBattlefield(manaLands(7, 0, 100)...), withTurn(9, 0, "precombat_main"))
	// The cycler is offered last, so a tie would take the other.
	in := input(0, v, passMove(0), cast(pilferer.InstanceID), cast(mako.InstanceID))

	if got := chose(t, in, decide(t, heuristic.New(), in)); got != in.Moves[2].Label {
		t.Fatalf("chose %q, want the Mako discarded", got)
	}
	pol, w := heuristic.New(), heuristic.DefaultWeights()
	// A 1/1 in hand is worth Power + Toughness; the cycler is floored at
	// one card, the card cycling it would find.
	gap := rankValue(t, pol, in, in.Moves[2].Label) - rankValue(t, pol, in, in.Moves[1].Label)
	if want := w.Power + w.Toughness - w.Hand; !nearly(gap, want) {
		t.Errorf("the Mako is %.3f cheaper to discard than the Pilferer, want %.3f", gap, want)
	}
	off := heuristic.NewWithConfig(noSelfReplacing())
	if gap := rankValue(t, off, in, in.Moves[2].Label) - rankValue(t, off, in, in.Moves[1].Label); !nearly(gap, 0) {
		t.Errorf("with the switch off the two 1/1s differ by %.3f, want 0", gap)
	}
}

// A big body with cycling keeps its price, less the mana the discard
// spares: SpellPerMana per mana of the cycling cost.
func TestABigCyclerLosesOnlyTheSparedMana(t *testing.T) {
	windfall, cast := windfallFixture(t)
	big := creature(cardID(2), 0, "Big Cycler", 6, 6)
	big.ManaCost = "{4}{G}{G}"
	cyc := withHandRow(big, "{2}{G}", &protocol.PurposeView{Draws: 1})
	cyc.InstanceID = cardID(3)
	price := func(pol *heuristic.Policy, c protocol.CardView) float64 {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, c)), newSeat(1)},
			withBattlefield(manaLands(7, 0, 100)...), withTurn(9, 0, "precombat_main"))
		in := input(0, v, passMove(0), cast(c.InstanceID))
		return rankValue(t, pol, in, in.Moves[1].Label)
	}
	gap := price(heuristic.New(), cyc) - price(heuristic.New(), big)
	if want := 3 * heuristic.DefaultConfig().SpellPerMana; !nearly(gap, want) {
		t.Errorf("cycling made the 6/6 %.3f cheaper to discard, want the spared {2}{G}'s %.3f", gap, want)
	}
	off := heuristic.NewWithConfig(noSelfReplacing())
	if gap := price(off, cyc) - price(off, big); !nearly(gap, 0) {
		t.Errorf("with the switch off cycling changed the discard by %.3f", gap)
	}
}

// Rows that are not "turn this card into a card" leave the price alone:
// a land with cycling (a land drop first), a discard-self row that
// declares no draw or tutor (channel), and one whose cost asks for more
// than mana.
func TestOnlyASelfReplacingNonlandRowIsCheaper(t *testing.T) {
	windfall, cast := windfallFixture(t)
	body := creature(cardID(2), 0, "Two Two", 2, 2)
	channel := withHandRow(body, "{1}{G}", &protocol.PurposeView{Tokens: 1})
	lifeCycler := cycler(body)
	lifeCycler.ZoneAbilities[0].LifeCost = 2
	tower := cycler(land(cardID(2), 0))
	plainLand := land(cardID(2), 0)
	price := func(c protocol.CardView) float64 {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, c)), newSeat(1)},
			withBattlefield(manaLands(3, 0, 100)...), withTurn(4, 0, "precombat_main"))
		in := input(0, v, passMove(0), cast(c.InstanceID))
		return rankValue(t, heuristic.New(), in, in.Moves[1].Label)
	}
	for _, tc := range []struct {
		name      string
		got, base protocol.CardView
	}{
		{"a cycling land", tower, plainLand},
		{"a channel row", channel, body},
		{"a row that also costs life", lifeCycler, body},
	} {
		if gap := price(tc.got) - price(tc.base); !nearly(gap, 0) {
			t.Errorf("%s changed the discard by %.3f, want 0", tc.name, gap)
		}
	}
}

// A discard the bot chooses (a loot, a discard prompt, the cleanup step)
// reads the same price: of two 1/1s, the cycler goes.
func TestADiscardChoicePitchesTheCycler(t *testing.T) {
	pilferer := creature(cardID(2), 0, "Impulsive Pilferer", 1, 1)
	mako := cycler(creature(cardID(3), 0, "Marauding Mako", 1, 1))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(pilferer, mako)), newSeat(1)},
		withBattlefield(sixLands()...), withTurn(5, 0, "cleanup"))
	// The cycler is offered last, so a tie would take the other.
	in := input(0, v, discardMove(t, 0, "discard the pilferer", pilferer.InstanceID), discardMove(t, 0, "discard the mako", mako.InstanceID))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard the mako" {
		t.Fatalf("chose %q, want the cycler", got)
	}
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(noSelfReplacing()), in)); got != "discard the pilferer" {
		t.Errorf("with the switch off it chose %q, want the first of two equal cards", got)
	}
}

func TestBaselineHasNoSelfReplacingDiscards(t *testing.T) {
	if heuristic.BaselineConfig().SelfReplacingDiscards {
		t.Error("BaselineConfig prices self-replacing discards; it is the pre-S66 heuristic")
	}
}
