package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cast_mana_test.go — ADR 0136 §2, owner answer 1: every cast move
// states the total mana it charges (CR 601.2f) on MoveCost.Mana, as the
// enumerator's payment check priced it: the commander tax (CR 903.8),
// X, the board's increases and reductions, and an alternative cost
// (CR 118.9). Each case also dispatches every move it checks and counts
// the lands the engine tapped for it, so the stamp is held to what the
// cast actually charges, not only to what the enumerator thinks.

const (
	oracleBlaze         = "0596920f-9946-42f4-a03b-24aab67f9f1b"
	oracleFaithlessLoot = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"
)

// castMovesOf returns the cast moves for one source card.
func castMovesOf(moves []legal.Move, src uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Source == src && m.Kind == legal.KindCast {
			out = append(out, m)
		}
	}
	return out
}

// tappedBy dispatches m against a clone of g and reports how many of
// the seat's untapped lands the cast tapped: with only basic lands and
// an empty pool, the mana the engine charged.
func tappedBy(t *testing.T, g *game.Game, seat uuid.UUID, m legal.Move) int {
	t.Helper()
	clone := g.Clone()
	untapped := func(gg *game.Game) int {
		n := 0
		for _, c := range gg.Battlefield.Cards {
			if c.Controller == seat && !c.Tapped && c.IsLand() {
				n++
			}
		}
		return n
	}
	before := untapped(clone)
	if err := actions.Dispatch(clone, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params,
	}); err != nil {
		t.Fatalf("move %q rejected: %v", m.Label, err)
	}
	return before - untapped(clone)
}

// wantCastMana checks every cast of src states `want`, and that the
// engine tapped `lands` lands to pay each one.
func wantCastMana(t *testing.T, g *game.Game, seat, src uuid.UUID, want string, lands int) []legal.Move {
	t.Helper()
	moves := legal.EnumerateFor(g, seat)
	casts := castMovesOf(moves, src)
	if len(casts) == 0 {
		t.Fatalf("no cast offered; moves: %v", labels(moves))
	}
	for _, m := range casts {
		if m.Cost == nil || m.Cost.Mana != want {
			t.Errorf("%q states cost %+v, want mana %q", m.Label, m.Cost, want)
		}
		if got := tappedBy(t, g, seat, m); got != lands {
			t.Errorf("%q: the engine tapped %d lands, the move states %q (%d)", m.Label, got, want, lands)
		}
	}
	return casts
}

// The commander tax: a commander cast once before costs {2} more, and
// the move says so.
func TestCastManaStatesTheCommanderTax(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	basicLands(g, active, 3, "Forest")
	basicLands(g, active, 3, "Island")
	id := uuid.New()
	active.Command.PushTop(game.Card{
		InstanceID: id, Name: "Tatyova Stand-in", TypeLine: "Legendary Creature — Dryad",
		ManaCost: "{2}{G}{U}", Power: 3, Toughness: 3,
		Owner: active.ID, Controller: active.ID, IsCommander: true,
		KnownBy: map[uuid.UUID]bool{active.ID: true},
	})
	active.CommanderCasts[id] = 1
	advanceTo(t, g, game.StepPrecombatMain)

	wantCastMana(t, g, active.ID, id, "{4}{G}{U}", 6)
}

// X: the move states its X settled into the generic component.
func TestCastManaSettlesX(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	basicLands(g, active, 4, "Mountain")
	blaze := handCard(active, game.Card{Name: "Blaze", TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: oracleBlaze})
	advanceTo(t, g, game.StepPrecombatMain)

	for _, m := range wantCastMana(t, g, active.ID, blaze, "{3}{R}", 4) {
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.XValue != 3 {
			t.Errorf("%q announces X=%d (%v), want 3", m.Label, p.XValue, err)
		}
	}
}

// A cost reduction: affinity for artifacts takes three off Thought
// Monitor's {6}{U}.
func TestCastManaStatesACostReduction(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	monitor := handCard(active, game.Card{Name: "Thought Monitor", TypeLine: "Artifact Creature — Construct", ManaCost: "{6}{U}", OracleID: oracleThoughtMonitor})
	basicLands(g, active, 1, "Island")
	basicLands(g, active, 3, "Plains")
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	advanceTo(t, g, game.StepPrecombatMain)

	wantCastMana(t, g, active.ID, monitor, "{3}{U}", 4)
}

// An alternative cost: Faithless Looting's flashback {2}{R} from the
// graveyard, not its printed {R}.
func TestCastManaStatesAnAlternativeCost(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	basicLands(g, active, 3, "Mountain")
	id := uuid.New()
	active.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{R}",
		OracleID: oracleFaithlessLoot, Owner: active.ID, Controller: active.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	wantCastMana(t, g, active.ID, id, "{2}{R}", 3)
}

// Nothing but casts gains a mana stamp: a land drop states no cost.
func TestCastManaIsOnCastsOnly(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	handCard(active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Kind != legal.KindCast && m.Cost != nil && m.Cost.Mana != "" {
			t.Errorf("%q (%s) states mana %q; only a cast or an attack tax does", m.Label, m.Kind, m.Cost.Mana)
		}
	}
}

// #2701, CR 601.2f and 118.7a: a generic reduction comes off the mana
// announced for X once the printed generic is gone. Blaze's {X}{R}
// under a {2} reduction with three Mountains is affordable at X = 4:
// four for X, less two, plus {R}. The move states {2}{R}, the engine
// taps three lands for it, and it still announces X = 4.
func TestCastManaTakesAReductionOffX(t *testing.T) {
	const reducer = "test-x-reducer"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withCostModifiers(t, reducer, []game.CostModifier{{
		Kind:   game.CostReduction,
		Label:  "Spells you cast cost {2} less to cast.",
		Amount: func(game.CostQuery) int { return 2 },
	}})
	battlefieldCard(g, active, game.Card{Name: "Test Reducer", TypeLine: "Artifact", OracleID: reducer})
	basicLands(g, active, 3, "Mountain")
	blaze := handCard(active, game.Card{Name: "Blaze", TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: oracleBlaze})
	advanceTo(t, g, game.StepPrecombatMain)

	for _, m := range wantCastMana(t, g, active.ID, blaze, "{2}{R}", 3) {
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.XValue != 4 {
			t.Errorf("%q announces X=%d (%v), want 4", m.Label, p.XValue, err)
		}
	}
}
