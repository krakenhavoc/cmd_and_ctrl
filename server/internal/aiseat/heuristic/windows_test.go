package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// windows_test.go is ADR 0126 §5: the bot's own last main phase and the
// end step just before its turn, where a move that costs only mana and
// taps needs only LeftoverThreshold.

// looter is a creature with a "{T}: Draw a card, then discard a card"
// row, Mary Read and Anne Bonny's shape.
func looter(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := creature(id, controller, "Looter", 3, 3, opts...)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Label: "{T}: Draw a card, then discard a card", TapCost: true}}
	return c
}

func lootMove(t *testing.T, seat int, id string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: "Loot", Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"source_card_id": id, "ability_index": 0}),
	}
}

// Four seats; the bot is seat 2, so seat 1's end step is the one just
// before its turn.
func fourSeats(hand ...protocol.CardView) []protocol.PlayerView {
	return []protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(hand...)), newSeat(3)}
}

func TestTheBotLootsInTheEndStepBeforeItsTurn(t *testing.T) {
	bf := withBattlefield(looter(cardID(20), 2))
	in := input(2, newView(fourSeats(), bf, withTurn(5, 1, "end")), passMove(2), lootMove(t, 2, cardID(20)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Loot" {
		t.Fatalf("chose %q in the end step before its turn, want the loot", got)
	}
	// The baseline is the pre-S66 heuristic: 0.20 against 1.50.
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), in)); got != "Pass priority" {
		t.Fatalf("baseline chose %q, want the pass", got)
	}
}

func TestTheBotDoesNotLootInAnotherSeatsEndStep(t *testing.T) {
	bf := withBattlefield(looter(cardID(20), 2))
	// Seat 0's end step: seat 1 plays next, and the looter would be
	// tapped through seat 1's turn.
	in := input(2, newView(fourSeats(), bf, withTurn(5, 0, "end")), passMove(2), lootMove(t, 2, cardID(20)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q in seat 0's end step, want the pass", got)
	}
}

func TestTheEndStepBeforeYoursSkipsAnEliminatedSeat(t *testing.T) {
	seats := fourSeats()
	seats[1] = newSeat(1, withEliminated())
	bf := withBattlefield(looter(cardID(20), 2))
	in := input(2, newView(seats, bf, withTurn(5, 0, "end")), passMove(2), lootMove(t, 2, cardID(20)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Loot" {
		t.Fatalf("chose %q; seat 1 is out, so seat 0's end step is the one before the bot's turn", got)
	}
}

func TestTheWindowWaitsForAnEmptyStack(t *testing.T) {
	bf := withBattlefield(looter(cardID(20), 2))
	theirs := spell(cardID(31), 1, "Their Spell", "{2}")
	in := input(2, newView(fourSeats(), bf, withStack(theirs), withTurn(5, 1, "end")), passMove(2), lootMove(t, 2, cardID(20)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q with a spell on the stack, want the pass", got)
	}
}

// An untargeted one-mana instant (Entomb's shape) is priced at
// SpellFloor, 0.10 above the card it costs, and is cast with leftover
// mana in the end step before the bot's turn. In the bot's own second
// main phase it waits for that step: the mana stays untapped until then.
// A one-mana sorcery (Preordain's shape) has no later window, so the
// second main phase is where it goes.
func TestACheapUntargetedSpellIsCastWithLeftoverMana(t *testing.T) {
	tutor := spell(cardID(1), 2, "Entomb", "{B}")
	cantrip := spell(cardID(2), 2, "Preordain", "{U}")
	cantrip.TypeLine = "Sorcery"
	bf := withBattlefield(land(cardID(10), 2))
	for _, tc := range []struct {
		name string
		turn viewOpt
		want string
	}{
		{"end step before the bot's turn", withTurn(5, 1, "end"), "Cast Entomb"},
		{"the bot's second main phase", withTurn(5, 2, "postcombat_main"), "Cast Preordain"},
		{"the bot's first main phase", withTurn(5, 2, "precombat_main"), "Pass priority"},
	} {
		moves := []legal.Move{passMove(2), castMove(t, 2, cardID(1), "Cast Entomb")}
		if tc.name != "end step before the bot's turn" {
			moves = append(moves, castMove(t, 2, cardID(2), "Cast Preordain"))
		}
		in := input(2, newView(fourSeats(tutor, cantrip), bf, tc.turn), moves...)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("%s: chose %q, want %q", tc.name, got, tc.want)
		}
	}
	// With only the instant on offer, the second main phase passes.
	in := input(2, newView(fourSeats(tutor), bf, withTurn(5, 2, "postcombat_main")), passMove(2), castMove(t, 2, cardID(1), "Cast Entomb"))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("second main phase, instant only: chose %q, want the pass (it waits for the end step)", got)
	}
}

// The floor does not crowd out development: a cheap cantrip still loses
// to a three-drop in the first main phase.
func TestTheCantripDoesNotBeatTheThreeDrop(t *testing.T) {
	cantrip := spell(cardID(1), 2, "Preordain", "{U}")
	bear := creature(cardID(2), 2, "Three Drop", 3, 3)
	bear.ManaCost = "{2}{G}"
	in := input(2, newView(fourSeats(cantrip, bear), withTurn(5, 2, "precombat_main")),
		passMove(2), castMove(t, 2, cardID(1), "Cast Preordain"), castMove(t, 2, cardID(2), "Cast Three Drop"))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Three Drop" {
		t.Fatalf("chose %q, want the three-drop", got)
	}
}

// A move that costs more than mana and taps keeps the normal bar: a
// sacrifice (Harrow's land), a discard. Each cast here is worth more
// than passing and less than InstantThreshold.
func TestALeftoverWindowNeedsAManaAndTapsCost(t *testing.T) {
	ramp := spell(cardID(1), 2, "Big Spell", "{4}{B}")
	bf := withBattlefield(land(cardID(10), 2))
	for name, extra := range map[string]map[string]any{
		"sacrifice": {"sacrifice_ids": []string{cardID(10)}},
		"discard":   {"discard_ids": []string{cardID(99)}},
	} {
		params := map[string]any{"instance_id": cardID(1), "from_zone": "hand"}
		for k, v := range extra {
			params[k] = v
		}
		m := legal.Move{Type: legal.TypeCastSpell, Player: seatID(2), Kind: legal.KindCast, Label: "Cast it",
			Source: uuid.MustParse(cardID(1)), Params: mustJSON(t, params)}
		in := input(2, newView(fourSeats(ramp), bf, withTurn(5, 1, "end")), passMove(2), m)
		if v := rankOf(t, in, "Cast it"); v <= 0 {
			t.Fatalf("%s: priced %+.2f; the test needs a cast worth more than passing", name, v)
		}
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("%s: chose %q, want the pass", name, got)
		}
	}
}

// A targeted spell gets no floor, so holding removal outside the
// windows is unchanged.
func TestTheSpellFloorIsForUntargetedSpells(t *testing.T) {
	bolt := spell(cardID(1), 2, "Lightning Bolt", "{R}")
	mouse := creature(cardID(20), 1, "Mouse", 1, 1)
	in := input(2, newView(fourSeats(bolt), withBattlefield(land(cardID(10), 2), mouse), withTurn(5, 0, "upkeep")),
		passMove(2), castMove(t, 2, cardID(1), "Bolt the mouse", cardTarget(cardID(20))))
	r := heuristic.New().Rank(t.Context(), in)
	b := heuristic.NewWithConfig(heuristic.BaselineConfig()).Rank(t.Context(), in)
	for i := range r {
		if r[i] != b[i] {
			t.Fatalf("the floor moved a targeted cast: %+v against the baseline's %+v", r[i], b[i])
		}
	}
}

// Tapping a creature that could attack, in the bot's own first main
// phase, costs the attack: the loot is not taken there.
func TestTappingAnAttackerBeforeCombatCostsTheAttack(t *testing.T) {
	bf := withBattlefield(looter(cardID(20), 2))
	pre := input(2, newView(fourSeats(), bf, withTurn(5, 2, "precombat_main")), passMove(2), lootMove(t, 2, cardID(20)))
	post := input(2, newView(fourSeats(), bf, withTurn(5, 2, "postcombat_main")), passMove(2), lootMove(t, 2, cardID(20)))
	vPre := rankOf(t, pre, "Loot")
	vPost := rankOf(t, post, "Loot")
	if vPre >= 0 {
		t.Errorf("loot before combat priced %+.2f, want below passing: the looter's attack is spent", vPre)
	}
	if vPost <= 0 {
		t.Errorf("loot after combat priced %+.2f, want above passing (0.50 less the blocker)", vPost)
	}
	// Above passing, but tapping the looter after combat spends a
	// blocker for every opponent's turn: the second main phase is a
	// leftover window for mana, not for that tap. The loot waits for the
	// end step before the bot's turn.
	if got := chose(t, post, decide(t, heuristic.New(), post)); got != "Pass priority" {
		t.Errorf("second main phase: chose %q, want the pass", got)
	}
}

func rankOf(t *testing.T, in aiseat.Input, label string) float64 {
	t.Helper()
	for _, c := range heuristic.New().Rank(t.Context(), in) {
		if in.Moves[c.Index].Label == label {
			return c.Value
		}
	}
	t.Fatalf("no move labelled %q", label)
	return 0
}
