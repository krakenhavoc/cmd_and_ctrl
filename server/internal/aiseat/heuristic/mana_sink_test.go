package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mana_sink_test.go — #4559: Lion Sash's "{W}: Exile target card from a
// graveyard" is a repeatable mana sink. The bot casts its spells first
// and spends what is left on the sink, and prices each exile by what it
// takes away. The whole-engine version,
// with the real Lion Sash, is aiseat/mana_sink_bot_test.go.

const sashRow = "{W}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this permanent."

// lionSash is seat 0's Lion Sash, a 1/1 creature with its exile row and
// the row's declared target entry, as the server projects it.
func lionSash(id string) protocol.CardView {
	c := creature(id, 0, "Lion Sash", 1, 1)
	c.TypeLine = "Artifact Creature — Equipment Cat"
	c.ManaCost = "{1}{W}"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: sashRow, ManaCost: "{W}",
		Purpose: &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{Slot: 0, Exiles: true, ExileGrows: "permanent"}}},
	}}
	return c
}

func sashMove(t *testing.T, src, target, name string) legal.Move {
	return activateMove(t, 0, src, "Lion Sash: exile "+name, &legal.MoveCost{Mana: "{W}"}, cardTarget(target))
}

// noSinkRule is today's heuristic with #4559 backed out.
func noSinkRule() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) {
		c.SinkManaLast = false
		c.PriceGraveyardExile = false
	})
}

// sinkBoard: Lion Sash, two Plains, and graveyards holding an opponent's
// creature, instant and land and the bot's own creature.
func sinkBoard(step string, active int, hand ...protocol.CardView) protocol.GameView {
	sash := lionSash(cardID(1))
	return newView([]protocol.PlayerView{
		newSeat(0, withHand(hand...), withGraveyard(creature(cardID(20), 0, "Own Angel", 4, 4))),
		newSeat(1, withGraveyard(
			creature(cardID(30), 1, "Opposing Titan", 6, 6),
			spell(cardID(31), 1, "Opposing Bolt", "{R}"),
			land(cardID(32), 1),
		)),
	}, withBattlefield(append(twoPlains(), sash)...), withTurn(5, active, step))
}

// TestManaSinkWaitsForTheTurnsSpells: in its own first main phase, with
// a cheap creature to cast and Lion Sash's exile on offer, the bot casts
// the creature. Backed out, the exile is priced above the creature and
// wins, which is the bug: the next window offers it again with the mana
// the creature needed.
func TestManaSinkWaitsForTheTurnsSpells(t *testing.T) {
	bear := creature(cardID(10), 0, "Squire", 1, 2)
	bear.ManaCost = "{1}{W}"
	v := sinkBoard("precombat_main", 0, bear)
	moves := []legal.Move{
		passMove(0),
		castMove(t, 0, cardID(10), "Cast Squire"),
		sashMove(t, cardID(1), cardID(30), "Opposing Titan"),
		sashMove(t, cardID(1), cardID(31), "Opposing Bolt"),
	}
	in := input(0, v, moves...)
	cast := rankValue(t, noSinkRule(), in, "Cast Squire")
	oldSink := rankValue(t, noSinkRule(), in, "Lion Sash: exile Opposing Titan")
	if !(oldSink > cast && cast > heuristic.DefaultConfig().PassThreshold) {
		t.Fatalf("fixture: want the backed-out sink (%.2f) above a castable creature (%.2f)", oldSink, cast)
	}
	if got := chose(t, in, decide(t, noSinkRule(), in)); got != "Lion Sash: exile Opposing Titan" {
		t.Errorf("back-out chose %q; want it to reproduce the sink", got)
	}
	d := decide(t, heuristic.New(), in)
	if got := chose(t, in, d); got != "Cast Squire" {
		t.Errorf("chose %q (%s); want the creature before the sink", got, d.Reason)
	}
	// With nothing else worth casting, the mana is the sink's.
	in = input(0, v, passMove(0), sashMove(t, cardID(1), cardID(30), "Opposing Titan"))
	if d := decide(t, heuristic.New(), in); d.Index != 1 || !strings.Contains(d.Reason, "mana sink last") {
		t.Errorf("chose %q (%s); want the sink with mana nothing else wants", chose(t, in, d), d.Reason)
	}
}

// TestManaSinkSpendsLeftoverManaByWhatTheExileTakes: in the end step
// before the bot's turn the sink is the leftover window's, and each pick
// is priced by what exiling it does. An opponent's creature card denies
// a reanimation target and grows the Sash; their land grows it; the
// bot's own creature costs the graveyard plans it might feed; their
// instant grows nothing and denies only the unseen recursion a card
// might feed, so it is the cheapest. Backed out, every opposing pick was
// the same flat price.
func TestManaSinkSpendsLeftoverManaByWhatTheExileTakes(t *testing.T) {
	v := sinkBoard("end", 1)
	titan := sashMove(t, cardID(1), cardID(30), "Opposing Titan")
	bolt := sashMove(t, cardID(1), cardID(31), "Opposing Bolt")
	mountain := sashMove(t, cardID(1), cardID(32), "Opposing Mountain")
	angel := sashMove(t, cardID(1), cardID(20), "Own Angel")
	in := input(0, v, passMove(0), bolt, angel, mountain, titan)
	pol := heuristic.New()
	if got := chose(t, in, decide(t, pol, in)); got != titan.Label {
		t.Errorf("chose %q; want the opponent's creature card", got)
	}
	vt, vm := rankValue(t, pol, in, titan.Label), rankValue(t, pol, in, mountain.Label)
	va, vb := rankValue(t, pol, in, angel.Label), rankValue(t, pol, in, bolt.Label)
	if !(vt > vm && vm > va && va > vb) {
		t.Errorf("exile prices titan %.2f, land %.2f, own angel %.2f, bolt %.2f; want them in that order", vt, vm, va, vb)
	}
	old := noSinkRule()
	flat := rankValue(t, old, in, bolt.Label)
	if vb >= flat || vt <= flat {
		t.Errorf("back-out prices every opposing pick at %.2f; want the instant (%.2f) below it and the creature (%.2f) above", flat, vb, vt)
	}
	if rankValue(t, old, in, titan.Label) != flat {
		t.Error("back-out: want every opposing pick at the same flat price, as before")
	}
}

// TestManaSinkGoesAfterAnotherLeftoverMove: in the end step before the
// bot's turn, a cheap instant that clears the leftover bar is cast
// before the sink, though the sink prices higher. The sink spends what
// the instant leaves.
func TestManaSinkGoesAfterAnotherLeftoverMove(t *testing.T) {
	opt := spell(cardID(11), 0, "Opt", "{U}")
	v := sinkBoard("end", 1, opt)
	sink := sashMove(t, cardID(1), cardID(30), "Opposing Titan")
	in := input(0, v, passMove(0), sink, castMove(t, 0, cardID(11), "Cast Opt"))
	if s, o := rankValue(t, heuristic.New(), in, sink.Label), rankValue(t, heuristic.New(), in, "Cast Opt"); !(s > o && o > 0) {
		t.Fatalf("fixture: want the sink (%.2f) above a positive instant (%.2f)", s, o)
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Opt" {
		t.Errorf("chose %q; want the instant before the sink", got)
	}
	if got := chose(t, in, decide(t, noSinkRule(), in)); got != sink.Label {
		t.Errorf("back-out chose %q; want the higher-priced sink", got)
	}
	in = input(0, v, passMove(0), sink)
	if d := decide(t, heuristic.New(), in); chose(t, in, d) != sink.Label || !strings.Contains(d.Reason, "mana sink last") {
		t.Errorf("chose %q (%s); want the sink with the leftover mana", chose(t, in, d), d.Reason)
	}
}

// TestManaSinkIsOnlyAManaOnlyRow: a row that also taps its source can be
// used once a turn, so it is not a sink and keeps its old price and bar.
// The baseline turns both of #4559's rules off.
func TestManaSinkIsOnlyAManaOnlyRow(t *testing.T) {
	sash := lionSash(cardID(1))
	tapRow := sash
	tapRow.ActivatedAbilities = []protocol.ActivatedAbilityView{sash.ActivatedAbilities[0]}
	tapRow.ActivatedAbilities[0].TapCost = true
	for _, tc := range []struct {
		name string
		src  protocol.CardView
	}{{"tap row", tapRow}} {
		t.Run(tc.name, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1, withGraveyard(creature(cardID(30), 1, "Opposing Titan", 6, 6)))},
				withBattlefield(append(twoPlains(), tc.src)...), withTurn(5, 0, "precombat_main"))
			m := sashMove(t, cardID(1), cardID(30), "Opposing Titan")
			in := input(0, v, passMove(0), m)
			d := decide(t, heuristic.New(), in)
			if strings.Contains(d.Reason, "mana sink") {
				t.Errorf("a %s was treated as a mana sink: %s", tc.name, d.Reason)
			}
		})
	}
	if heuristic.BaselineConfig().SinkManaLast || heuristic.BaselineConfig().PriceGraveyardExile {
		t.Error("baseline enables #4559's mana-sink rules")
	}
}

// TestManaSinkWaitsForTheMainPhase: in the bot's own upkeep, stack empty,
// an exile priced above InstantThreshold would tap a Plains that stays
// tapped through the main phase. It waits for the main phase, where the
// spells go first. Backed out, it is taken in the upkeep.
func TestManaSinkWaitsForTheMainPhase(t *testing.T) {
	v := sinkBoard("upkeep", 0)
	titan := sashMove(t, cardID(1), cardID(30), "Opposing Titan")
	in := input(0, v, passMove(0), titan)
	if got := rankValue(t, heuristic.New(), in, titan.Label); got <= heuristic.DefaultConfig().InstantThreshold {
		t.Fatalf("fixture: want the exile (%.2f) above InstantThreshold", got)
	}
	if d := decide(t, heuristic.New(), in); d.Index != 0 || !strings.Contains(d.Reason, "wait for the main phase") {
		t.Errorf("chose %q (%s); want to wait for the main phase", chose(t, in, d), d.Reason)
	}
	cfg := heuristic.DefaultConfig()
	cfg.SinkManaLast = false
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(cfg), in)); got != titan.Label {
		t.Errorf("SinkManaLast off chose %q; want the exile in the upkeep", got)
	}
}
