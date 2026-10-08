package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_mana_value_cost_test.go — #2190: "Discard a card with mana
// value X: Counter target spell with mana value X." The discarded
// card's mana value IS the announced X, so the enumerator offers one
// payment per distinct mana value in the hand, each paired only with
// the spells of that value, with x_value the value. Every move is one
// the engine accepts (#544).

const oracleKozilekTheGreatDistortion = "4c1c1537-e519-4e2f-9bc2-d34b289d4487"

type manaValueDiscardPayload struct {
	XValue     int      `json:"x_value"`
	DiscardIDs []string `json:"discard_ids"`
	Targets    []struct {
		ID string `json:"id"`
	} `json:"targets"`
}

// kozilekMoves seats Kozilek and a hand of mana values 0, 1, 3, 3 and
// 4 on the active seat, puts one spell of each cost in `spellCosts` on
// the stack, and returns the seat's Kozilek activations.
func kozilekMoves(t *testing.T, spellCosts ...string) (*game.Game, *game.Player, []legal.Move, []manaValueDiscardPayload, map[string]uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	kozilek := battlefieldCard(g, seat, game.Card{
		Name: "Kozilek, the Great Distortion", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: oracleKozilekTheGreatDistortion, Power: 12, Toughness: 12,
	})
	spells := map[string]uuid.UUID{}
	for _, cost := range spellCosts {
		id := handCard(seat, game.Card{Name: "Seed " + cost, TypeLine: "Instant", ManaCost: cost})
		if err := g.CastSpell(seat.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("seed the stack with %s: %v", cost, err)
		}
		spells[cost] = id
	}
	hand := map[string]uuid.UUID{}
	for _, c := range []struct{ name, cost string }{
		{"Land", ""}, {"One", "{G}"}, {"Three A", "{2}{G}"}, {"Three B", "{1}{G}{G}"}, {"Four", "{3}{G}"},
	} {
		hand[c.name] = handCard(seat, game.Card{Name: c.name, TypeLine: "Instant", ManaCost: c.cost})
	}
	moves := legal.EnumerateFor(g, seat.ID)
	var payloads []manaValueDiscardPayload
	var kozMoves []legal.Move
	for _, m := range moves {
		if m.Kind != legal.KindActivate || m.Source != kozilek {
			continue
		}
		var p manaValueDiscardPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		payloads = append(payloads, p)
		kozMoves = append(kozMoves, m)
	}
	for k, v := range hand {
		spells["hand:"+k] = v
	}
	return g, seat, kozMoves, payloads, spells
}

// One spell of mana value 3: one move, whichever of the two
// mana-value-3 cards stands for the value, x_value 3, aimed at the
// spell.
func TestKozilekOffersOnePaymentPerDistinctManaValue(t *testing.T) {
	g, seat, moves, got, ids := kozilekMoves(t, "{2}{R}")
	if len(got) != 1 {
		t.Fatalf("offered %d Kozilek activations, want exactly 1: %+v", len(got), got)
	}
	p := got[0]
	if p.XValue != 3 || len(p.DiscardIDs) != 1 {
		t.Fatalf("payload %+v, want x_value 3 with one card named", p)
	}
	if p.DiscardIDs[0] != ids["hand:Three A"].String() && p.DiscardIDs[0] != ids["hand:Three B"].String() {
		t.Errorf("discards %s, want one of the mana value 3 cards", p.DiscardIDs[0])
	}
	if len(p.Targets) != 1 || p.Targets[0].ID != ids["{2}{R}"].String() {
		t.Errorf("targets %+v, want the spell on the stack", p.Targets)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Spells of different values each get their own matching card, and no
// card is paired with a spell of another value.
func TestKozilekPairsEachSpellWithACardOfItsOwnValue(t *testing.T) {
	g, seat, moves, got, ids := kozilekMoves(t, "{G}", "{3}{R}", "{6}")
	want := map[string]string{ // spell id -> the hand card standing for its value
		ids["{G}"].String():    ids["hand:One"].String(),
		ids["{3}{R}"].String(): ids["hand:Four"].String(),
	}
	if len(got) != len(want) {
		t.Fatalf("offered %d activations, want %d (the mana value 6 spell has no answer): %+v", len(got), len(want), got)
	}
	for _, p := range got {
		if len(p.Targets) != 1 || len(p.DiscardIDs) != 1 {
			t.Fatalf("payload %+v: want one target and one card", p)
		}
		if w, ok := want[p.Targets[0].ID]; !ok || p.DiscardIDs[0] != w {
			t.Errorf("payload %+v pairs the wrong card with the spell", p)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Nothing in the hand matches the spell: no activation at all (#544).
func TestKozilekNotOfferedWhenNoCardMatches(t *testing.T) {
	_, _, _, got, _ := kozilekMoves(t, "{5}{R}")
	if len(got) != 0 {
		t.Errorf("offered %d activations against a mana value 6 spell: %+v", len(got), got)
	}
}

// With nothing on the stack there is nothing to counter.
func TestKozilekNotOfferedWithAnEmptyStack(t *testing.T) {
	_, _, _, got, _ := kozilekMoves(t)
	if len(got) != 0 {
		t.Errorf("offered %d activations with an empty stack: %+v", len(got), got)
	}
}

// A Fireball cast for X=3 is mana value 4 while it is on the stack
// (CR 202.3e), so the four is the card that answers it.
func TestKozilekReadsTheXOfASpellOnTheStack(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	kozilek := battlefieldCard(g, seat, game.Card{
		Name: "Kozilek, the Great Distortion", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: oracleKozilekTheGreatDistortion, Power: 12, Toughness: 12,
	})
	fireball := handCard(seat, game.Card{Name: "Fireball", TypeLine: "Instant", ManaCost: "{X}{R}"})
	if err := g.CastSpell(seat.ID, fireball, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast Fireball: %v", err)
	}
	one := handCard(seat, game.Card{Name: "One", TypeLine: "Instant", ManaCost: "{G}"})
	four := handCard(seat, game.Card{Name: "Four", TypeLine: "Instant", ManaCost: "{3}{G}"})
	var got []manaValueDiscardPayload
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if m.Kind != legal.KindActivate || m.Source != kozilek {
			continue
		}
		var p manaValueDiscardPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		got = append(got, p)
	}
	if len(got) != 1 || got[0].XValue != 4 || got[0].DiscardIDs[0] != four.String() {
		t.Fatalf("payloads %+v, want exactly the mana value 4 card (not %s) at x_value 4", got, one)
	}
}
