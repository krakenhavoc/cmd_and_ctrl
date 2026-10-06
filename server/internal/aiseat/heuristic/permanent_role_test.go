package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// permanent_role_test.go is ADR 0126 §3 (PR 4): a permanent that is
// neither a creature nor a mana source is priced by its mana value and
// by its ability rows, and a creature adds its rows to its body.

func rows(kinds ...string) cardOpt {
	return func(c *protocol.CardView) {
		for _, k := range kinds {
			c.AbilityRows = append(c.AbilityRows, protocol.AbilityRowView{Kind: k, Label: k})
		}
	}
}

func unimplemented() cardOpt { return func(c *protocol.CardView) { c.Unimplemented = true } }

func enchantment(id string, controller int, name, cost string, opts ...cardOpt) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id,
		Name:       name,
		Owner:      seatID(controller).String(),
		Controller: seatID(controller).String(),
		TypeLine:   "Enchantment",
		ManaCost:   cost,
		KnownByYou: true,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// boardWith is the board value w gives seat 0 for exactly these
// permanents.
func boardWith(w heuristic.Weights, cards ...protocol.CardView) float64 {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(cards...))
	return w.Evaluate(v)[seatID(0).String()].Board
}

func TestUtilityPermanentsArePricedByManaValueAndRows(t *testing.T) {
	w := heuristic.DefaultWeights()
	cases := []struct {
		name string
		card protocol.CardView
		want float64
	}{
		// The worked examples in ADR 0126 §3, before the card's 1.20.
		{"Rhystic Study, MV 3, one triggered row", enchantment(cardID(1), 0, "Rhystic Study", "{2}{U}", rows("triggered")), 1.5 + 0.6},
		{"Impact Tremors, MV 2: the flat Permanent is the floor", enchantment(cardID(1), 0, "Impact Tremors", "{1}{R}", rows("triggered")), 1.2 + 0.6},
		{"a static row", enchantment(cardID(1), 0, "Boots", "{2}", rows("static")), 1.2 + 0.5},
		{"a big blank artifact", enchantment(cardID(1), 0, "Monolith", "{6}"), 3.0},
		{"rows are capped at three", enchantment(cardID(1), 0, "Engine", "{2}", rows("triggered", "triggered", "triggered", "triggered")), 1.2 + 3*0.6},
		{"an unimplemented card keeps the flat price", enchantment(cardID(1), 0, "Mystery", "{8}", unimplemented(), rows("triggered")), 1.2},
		{"a token has no mana value", enchantment(cardID(1), 0, "Clue", "", rows("activated")), 1.2 + 0.4},
	}
	for _, tc := range cases {
		if got := boardWith(w, tc.card); !near(got, tc.want) {
			t.Errorf("%s: board %.4f, want %.4f", tc.name, got, tc.want)
		}
	}
}

func TestCreatureRowsCountOnTheBoardButNotInCombat(t *testing.T) {
	w := heuristic.DefaultWeights()
	vanilla := creature(cardID(1), 0, "Bird", 1, 1)
	seer := creature(cardID(2), 0, "Viscera Seer", 1, 1, rows("activated"))
	if got, want := w.CreatureValue(&seer)-w.CreatureValue(&vanilla), 0.4; !near(got, want) {
		t.Errorf("an activated row adds %.4f to CreatureValue, want %.4f", got, want)
	}
	// Owner decision 3: CombatValue stays body-only in S66.
	if w.CombatValue(&seer) != w.CombatValue(&vanilla) {
		t.Errorf("CombatValue counts ability rows: %.4f vs %.4f", w.CombatValue(&seer), w.CombatValue(&vanilla))
	}
	// The rows are added before the summoning-sick multiplier.
	sickSeer := creature(cardID(2), 0, "Viscera Seer", 1, 1, rows("activated"), sick())
	if got, want := w.CreatureValue(&sickSeer), 0.9*(1.45+0.4); !near(got, want) {
		t.Errorf("a sick Seer is worth %.4f, want %.4f", got, want)
	}
	// An unimplemented creature's rows (it has none on the wire, but a
	// stale one must not count) add nothing.
	ghost := creature(cardID(3), 0, "Ghost", 1, 1, rows("triggered"), unimplemented())
	if w.CreatureValue(&ghost) != w.CreatureValue(&vanilla) {
		t.Errorf("an unimplemented creature's rows were counted")
	}
}

// The new terms are zero in BaselineConfig, so the baseline prices
// every permanent as the pre-S66 heuristic did.
func TestBaselineConfigPricesPermanentsFlat(t *testing.T) {
	w := heuristic.BaselineConfig().Weights
	study := enchantment(cardID(1), 0, "Rhystic Study", "{2}{U}", rows("triggered"))
	if got := boardWith(w, study); !near(got, w.Permanent) {
		t.Errorf("baseline Rhystic Study is %.4f, want the flat %.4f", got, w.Permanent)
	}
	seer := creature(cardID(2), 0, "Viscera Seer", 1, 1, rows("activated"))
	vanilla := creature(cardID(3), 0, "Bird", 1, 1)
	if w.CreatureValue(&seer) != w.CreatureValue(&vanilla) {
		t.Errorf("baseline counts a creature's rows")
	}
}

// The cast price reads the same function, so Rhystic Study clears the
// main-phase bar it used to sit exactly at.
func TestCastRhysticStudyInTheMainPhase(t *testing.T) {
	study := enchantment(cardID(10), 0, "Rhystic Study", "{2}{U}", rows("triggered"))
	v := newView(
		[]protocol.PlayerView{newSeat(0, withHand(study)), newSeat(1)},
		withBattlefield(land(cardID(1), 0), land(cardID(2), 0), land(cardID(3), 0)),
		withTurn(3, 0, "precombat_main"),
	)
	in := input(0, v, passMove(0), castMove(t, 0, cardID(10), "Cast Rhystic Study"))
	if d := decide(t, heuristic.New(), in); d.Index != 1 {
		t.Fatalf("the heuristic chose %d (%s), want Cast Rhystic Study", d.Index, d.Reason)
	}
	if d := decide(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), in); d.Index != 0 {
		t.Fatalf("the baseline chose %d (%s), want the pre-S66 pass", d.Index, d.Reason)
	}
}

// equip-before-attacking (ADR 0126 §8's equip position). The curated
// decks hold no Equipment, so no arena window can be harvested for it;
// this pins the same behaviour on a hand-built board. An Equipment is
// now worth its mana-value floor on the battlefield, and that must not
// make the bot keep it unattached: equipping the creature that is about
// to attack still beats passing.
func TestEquipBeforeAttacking(t *testing.T) {
	sword := equipment(cardID(20), 0, "Sword", func(c *protocol.CardView) { c.ManaCost = "{3}" })
	bear := creature(cardID(21), 0, "Bear", 2, 2)
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(sword, bear, land(cardID(1), 0), land(cardID(2), 0)),
		withTurn(5, 0, "precombat_main"),
	)
	in := input(0, v, passMove(0), activateMove(t, 0, cardID(20), "Sword: Equip {2}", nil, cardTarget(cardID(21))))
	for name, p := range map[string]*heuristic.Policy{
		"heuristic": heuristic.New(),
		"baseline":  heuristic.NewWithConfig(heuristic.BaselineConfig()),
	} {
		if d := decide(t, p, in); d.Index != 1 {
			t.Errorf("%s chose %d (%s), want the equip", name, d.Index, d.Reason)
		}
	}
}
