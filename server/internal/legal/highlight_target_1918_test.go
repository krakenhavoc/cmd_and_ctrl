package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// highlight_target_1918_test.go — #1918. The client's playable
// highlight is the enumerator's cast moves and nothing else
// (legalActions.ts castableFrom over the frame's legal_actions), so a
// spell lights up exactly when this package offers a cast for it.
// CR 601.2c: a spell that needs a target can't be cast without a legal
// one. These pin that for the cases the report named, and for the
// three shapes that ARE castable with nothing to point at: an
// overloaded spell (CR 702.96a turns "target" into "each"), a modal
// spell with one mode that has a target (CR 700.2), and an "up to one"
// clause (CR 115.1, 601.2c: zero targets is a legal choice).

const (
	oracleMurder              = "938b4e2c-88d9-4637-bc00-e228920c9a78"
	oracleCounterflux         = "8c983da8-1436-4c06-af9c-91b72cab48c1"
	oracleDispel              = "6d7be242-a072-40ce-b540-95880506cccd"
	oracleHeritageReclamation = "09955b4b-6052-4c27-8b63-9548482b3c5c"
)

type castProbe1918 struct {
	AlternativeCost string            `json:"alternative_cost"`
	Modes           []int             `json:"modes"`
	Targets         []json.RawMessage `json:"targets"`
}

func castsOf1918(t *testing.T, moves []legal.Move, src uuid.UUID) []castProbe1918 {
	t.Helper()
	var out []castProbe1918
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p castProbe1918
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

// Murder with no creature anywhere is not offered; one creature, and it
// is.
func TestRemovalNeedsALegalCreature(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	murder := handCard(active, game.Card{Name: "Murder", TypeLine: "Instant", ManaCost: "{1}{B}{B}", OracleID: oracleMurder})
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, basic("Swamp", "Swamp"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	if got := castsOf1918(t, legal.EnumerateFor(g, active.ID), murder); len(got) != 0 {
		t.Fatalf("Murder offered with no creature on the battlefield: %+v", got)
	}

	battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := castsOf1918(t, moves, murder)
	if len(got) != 1 || len(got[0].Targets) != 1 {
		t.Fatalf("want one Murder cast at the Bear, got %+v", got)
	}
}

// Dispel with nothing on the stack is not offered. Counterflux is, but
// only for its overload cost: overloaded it has no target (CR 702.96a),
// so an empty stack doesn't stop the cast. That is the card the report's
// screenshot shows lit.
func TestCounterspellsWithAnEmptyStack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	dispel := handCard(active, game.Card{Name: "Dispel", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleDispel})
	flux := handCard(active, game.Card{Name: "Counterflux", TypeLine: "Instant", ManaCost: "{U}{U}{R}", OracleID: oracleCounterflux})
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	if got := castsOf1918(t, moves, dispel); len(got) != 0 {
		t.Errorf("Dispel offered with an empty stack: %+v", got)
	}
	got := castsOf1918(t, moves, flux)
	if len(got) != 1 || got[0].AlternativeCost != "overload" || len(got[0].Targets) != 0 {
		t.Errorf("want exactly the overloaded Counterflux with no target, got %+v", got)
	}
}

// Heritage Reclamation over a board with no artifact, no enchantment and
// empty graveyards: the two destroy modes have nothing to target, the
// third's "up to one target card" takes zero, so the card is castable
// through that mode alone.
func TestModalUpToOneCastableWithNoTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	hr := handCard(active, game.Card{Name: "Heritage Reclamation", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: oracleHeritageReclamation})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	got := castsOf1918(t, moves, hr)
	if len(got) != 1 || len(got[0].Modes) != 1 || got[0].Modes[0] != 2 || len(got[0].Targets) != 0 {
		t.Errorf("want one cast choosing mode 2 with no target, got %+v", got)
	}
}
