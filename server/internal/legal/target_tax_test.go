package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// target_tax_test.go — #3860. Terror of the Peaks: "Spells your
// opponents cast that target this creature cost an additional 3 life to
// cast" (CR 601.2f, game/target_tax.go). The enumerator prices each
// target set with the same accessor CastSpell charges, drops the casts the
// seat cannot pay for (a player who can't pay can't target it), and tells
// the bot the price in MoveCost so it can weigh it.

const (
	oracleTerror = "f8e17f4f-080d-4bba-bd05-ca27e94ccecc"
	oracleBolt   = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// boltTable seats a Lightning Bolt and a Mountain with the active seat at
// `life`, an opponent's Terror of the Peaks and an opponent's Bear.
func boltTable(t *testing.T, life int) (g *game.Game, seat *game.Player, bolt, terror, bear uuid.UUID) {
	t.Helper()
	g = newTable(t)
	seat = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	seat.Life = life
	battlefieldCard(g, seat, basic("Mountain", "Mountain"))
	terror = battlefieldCard(g, opp, game.Card{
		Name: "Terror of the Peaks", TypeLine: "Creature — Dragon", ManaCost: "{3}{R}{R}",
		OracleID: oracleTerror, Power: 5, Toughness: 4,
	})
	bear = battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
	bolt = handCard(seat, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleBolt, Layout: "normal"})
	return g, seat, bolt, terror, bear
}

// boltAt is the Bolt casts that name `name`, by their label.
func boltAt(moves []legal.Move, bolt uuid.UUID, name string) []legal.Move {
	var out []legal.Move
	for _, m := range castMovesFor(moves, bolt) {
		if strings.Contains(m.Label, name) {
			out = append(out, m)
		}
	}
	return out
}

func TestEnumeratorPricesTheTaxOnEachTarget(t *testing.T) {
	g, seat, bolt, _, _ := boltTable(t, 20)
	moves := legal.EnumerateFor(g, seat.ID)

	atTerror := boltAt(moves, bolt, "Terror of the Peaks")
	if len(atTerror) != 1 {
		t.Fatalf("a Bolt at Terror of the Peaks is offered once, got %d: %v", len(atTerror), labels(castMovesFor(moves, bolt)))
	}
	m := atTerror[0]
	if m.Cost == nil || m.Cost.Life != 3 || m.Cost.TargetTaxLife != 3 || m.Cost.PhyrexianLife != 0 {
		t.Errorf("cost = %+v, want Life 3 of which 3 is the target tax", m.Cost)
	}
	if !strings.Contains(m.Label, "paying 3 life for what it targets") {
		t.Errorf("label %q should say what the target costs", m.Label)
	}

	atBear := boltAt(moves, bolt, "Bear")
	if len(atBear) != 1 {
		t.Fatalf("a Bolt at the Bear is offered once, got %d", len(atBear))
	}
	if c := atBear[0].Cost; c != nil && (c.Life != 0 || c.TargetTaxLife != 0) {
		t.Errorf("an untaxed target costs no life, got %+v", c)
	}
	dispatchAll(t, g, seat.ID, castMovesFor(moves, bolt))
}

// A seat that can't pay is not offered the cast at the taxed target
// (CR 119.4: exactly its life total is fine) -- and is still offered
// every other one.
func TestEnumeratorDropsACastTheSeatCannotPayTheTaxFor(t *testing.T) {
	for _, tc := range []struct {
		name string
		life int
		want bool
	}{
		{"2 life cannot pay 3", 2, false},
		{"3 life pays exactly its life total", 3, true},
		{"20 life", 20, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, seat, bolt, _, _ := boltTable(t, tc.life)
			moves := legal.EnumerateFor(g, seat.ID)
			if got := len(boltAt(moves, bolt, "Terror of the Peaks")) > 0; got != tc.want {
				t.Errorf("Bolt at Terror offered = %v, want %v", got, tc.want)
			}
			if len(boltAt(moves, bolt, "Bear")) == 0 {
				t.Error("the untaxed target is offered whatever the life total")
			}
			// Soundness (#544): everything offered dispatches.
			dispatchAll(t, g, seat.ID, castMovesFor(moves, bolt))
		})
	}
}

// Its controller's own spells are free to name it.
func TestEnumeratorDoesNotTaxTheDragonsController(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, basic("Mountain", "Mountain"))
	battlefieldCard(g, seat, game.Card{
		Name: "Terror of the Peaks", TypeLine: "Creature — Dragon", ManaCost: "{3}{R}{R}",
		OracleID: oracleTerror, Power: 5, Toughness: 4,
	})
	bolt := handCard(seat, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleBolt, Layout: "normal"})
	moves := legal.EnumerateFor(g, seat.ID)
	at := boltAt(moves, bolt, "Terror of the Peaks")
	if len(at) != 1 {
		t.Fatalf("Bolt at your own Dragon: got %d moves", len(at))
	}
	if c := at[0].Cost; c != nil && c.TargetTaxLife != 0 {
		t.Errorf("its controller pays nothing: %+v", c)
	}
}

// The tax shares the life total with the cast's other life costs: a
// Dismember paying both Phyrexian symbols with life (4) plus the tax (3)
// must fit the life total together (CR 119.4), not each on its own.
func TestEnumeratorHoldsTheTaxBackFromPhyrexianLife(t *testing.T) {
	for _, tc := range []struct {
		name       string
		life       int
		wantTerror bool
	}{
		{"4 life: the symbols fit, the tax on top does not", 4, false},
		{"6 life: still short of 4 + 3", 6, false},
		{"7 life pays 4 for the symbols and 3 for the tax", 7, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			seat := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			clearHand(seat)
			advanceTo(t, g, game.StepPrecombatMain)
			seat.Life = tc.life
			// {1} from the Mountain; both {B/P} can only be paid with life.
			battlefieldCard(g, seat, basic("Mountain", "Mountain"))
			battlefieldCard(g, opp, game.Card{
				Name: "Terror of the Peaks", TypeLine: "Creature — Dragon", ManaCost: "{3}{R}{R}",
				OracleID: oracleTerror, Power: 5, Toughness: 4,
			})
			battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
			dismember := handCard(seat, game.Card{
				Name: "Dismember", TypeLine: "Instant", ManaCost: "{1}{B/P}{B/P}",
				OracleID: "fd74f8eb-0253-42dd-8277-186d4934da38", Layout: "normal",
			})
			moves := legal.EnumerateFor(g, seat.ID)
			got := len(boltAt(moves, dismember, "Terror of the Peaks")) > 0
			if got != tc.wantTerror {
				t.Errorf("Dismember at Terror offered = %v, want %v (%v)", got, tc.wantTerror, labels(castMovesFor(moves, dismember)))
			}
			if tc.life >= 4 && len(boltAt(moves, dismember, "Bear")) == 0 {
				t.Errorf("Dismember at the untaxed Bear is offered at %d life", tc.life)
			}
			dispatchAll(t, g, seat.ID, castMovesFor(moves, dismember))
		})
	}
}
