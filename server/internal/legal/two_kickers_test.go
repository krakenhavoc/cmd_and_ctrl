package legal_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// two_kickers_test.go — #2153's enumerator half, in the #544 agreement
// style: a card with two kicker costs (CR 702.33b) is offered unkicked,
// kicked with either, and kicked with both — each announcement one the
// engine accepts, and none the seat cannot pay for.

const oracleThornscapeBattlemage = "8d4d6806-cb01-49e3-91cc-fb0e7f7a8684"

// kickerSets reads the optional_costs announcement of every cast the
// enumerator offered for `src`, as strings so the set compares.
func kickerSets(t *testing.T, moves []legal.Move, src uuid.UUID) []string {
	t.Helper()
	var out []string
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p struct {
			OptionalCosts []int `json:"optional_costs"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		s := fmt.Sprint(p.OptionalCosts)
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	slices.Sort(out)
	return out
}

func thornscapeInHand(g *game.Game, p *game.Player) uuid.UUID {
	return handCard(p, game.Card{
		Name: "Thornscape Battlemage", TypeLine: "Creature — Elf Wizard",
		OracleID: oracleThornscapeBattlemage, ManaCost: "{2}{G}", Power: 2, Toughness: 2,
	})
}

// TestTwoKickerCostsAreEnumeratedTogether: with mana for everything,
// all four announcements are offered and every one is accepted.
func TestTwoKickerCostsAreEnumeratedTogether(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Forest")
	basicLands(g, seat, 1, "Mountain")
	basicLands(g, seat, 1, "Plains")
	mage := thornscapeInHand(g, seat)

	moves := castMovesFor(legal.EnumerateFor(g, seat.ID), mage)
	want := []string{"[0 1]", "[0]", "[1]", "[]"}
	if got := kickerSets(t, moves, mage); !slices.Equal(got, want) {
		t.Errorf("announced kicker sets = %v, want %v: %v", got, want, labels(moves))
	}
	if !hasLabel(moves, "Cast Thornscape Battlemage (Kicker {R}, Kicker {W})") {
		t.Errorf("the doubly kicked line is missing or mislabelled: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// TestTwoKickerLinesNeedTheirOwnMana: with only {R} to spare, the {W}
// line and the both line are not offered — never a cast the engine
// would refuse for mana.
func TestTwoKickerLinesNeedTheirOwnMana(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Forest")
	basicLands(g, seat, 1, "Mountain")
	mage := thornscapeInHand(g, seat)

	moves := castMovesFor(legal.EnumerateFor(g, seat.ID), mage)
	want := []string{"[0]", "[]"}
	if got := kickerSets(t, moves, mage); !slices.Equal(got, want) {
		t.Errorf("announced kicker sets = %v, want %v: %v", got, want, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}
