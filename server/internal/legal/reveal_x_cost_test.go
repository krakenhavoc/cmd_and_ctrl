package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// reveal_x_cost_test.go — #2598: "Reveal X <colour> cards from your
// hand" announces its count as X (CR 602.2b), so the enumerator offers
// a bounded ladder of reveals — the first 1, 2, 3 matching cards — each
// with x_value equal to the cards named and a target count that agrees
// with it, never the X = 0 no-op for a card whose whole effect is X
// (#810), and nothing when the hand holds no matching card. Every move
// is one the engine accepts (#544).

const oracleMartyrOfBones = "a099d8fa-d51e-4dbc-a03f-6f912097d41e"

type revealXPayload struct {
	XValue    int      `json:"x_value"`
	RevealIDs []string `json:"reveal_ids"`
	Targets   []struct {
		ID string `json:"id"`
	} `json:"targets"`
}

// martyrBonesMoves seats a Martyr of Bones with `black` black cards and
// two white ones in hand, and three cards in an opposing graveyard.
func martyrBonesMoves(t *testing.T, black int) (*game.Game, *game.Player, []legal.Move, []revealXPayload) {
	t.Helper()
	return martyrBonesMovesBudget(t, black, 0)
}

// martyrBonesMovesBudget is martyrBonesMoves with a wider
// MaxExpansionPerSource, so the target walk reaches the multi-card sets
// the default budget cuts off.
func martyrBonesMovesBudget(t *testing.T, black, budget int) (*game.Game, *game.Player, []legal.Move, []revealXPayload) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Swamp")
	martyr := battlefieldCard(g, seat, game.Card{
		Name: "Martyr of Bones", TypeLine: "Creature — Human Wizard", ManaCost: "{B}",
		OracleID: oracleMartyrOfBones, Power: 1, Toughness: 1,
	})
	for i := 0; i < black; i++ {
		handCard(seat, game.Card{Name: "Black Card", TypeLine: "Sorcery", ManaCost: "{1}{B}"})
	}
	for i := 0; i < 2; i++ {
		handCard(seat, game.Card{Name: "White Card", TypeLine: "Sorcery", ManaCost: "{1}{W}"})
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		graveyardCard(opp, game.Card{Name: "Their Card", TypeLine: "Sorcery"})
	}
	var moves []legal.Move
	if budget > 0 {
		moves = legal.EnumerateForWithOptions(g, seat.ID, legal.Options{MaxExpansionPerSource: budget})
	} else {
		moves = legal.EnumerateFor(g, seat.ID)
	}
	var payloads []revealXPayload
	for _, m := range moves {
		if m.Kind != legal.KindActivate || m.Source != martyr {
			continue
		}
		var p revealXPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		payloads = append(payloads, p)
	}
	return g, seat, moves, payloads
}

// With five black cards the ladder is 1, 2, 3: never 0, never past the
// cap, x_value always the cards named, every revealed card black, and a
// target set never larger than X. Every move dispatches.
func TestRevealXOffersALadderOfCounts(t *testing.T) {
	g, seat, moves, got := martyrBonesMoves(t, 5)
	counts := map[int]bool{}
	for _, p := range got {
		if p.XValue != len(p.RevealIDs) {
			t.Errorf("x_value %d does not match the %d cards named", p.XValue, len(p.RevealIDs))
		}
		if len(p.Targets) > p.XValue {
			t.Errorf("%d targets at X=%d", len(p.Targets), p.XValue)
		}
		counts[len(p.RevealIDs)] = true
		for _, id := range p.RevealIDs {
			c := cardInHand(t, seat, id)
			if c.Name != "Black Card" {
				t.Errorf("revealed %q for a black reveal", c.Name)
			}
		}
	}
	for _, n := range []int{1, 2, 3} {
		if !counts[n] {
			t.Errorf("count %d not offered (offered %v)", n, counts)
		}
	}
	if counts[0] || counts[4] || counts[5] {
		t.Errorf("counts offered outside the ladder: %v", counts)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Some move at the top of the ladder names targets, and one names none
// (an "up to X" clause is never unfillable).
func TestRevealXOffersTargetsUpToX(t *testing.T) {
	_, _, _, got := martyrBonesMoves(t, 3)
	var withTargets, without bool
	for _, p := range got {
		if len(p.Targets) > 0 {
			withTargets = true
		} else {
			without = true
		}
	}
	if !withTargets || !without {
		t.Errorf("offered with targets = %v, without = %v; want both", withTargets, without)
	}
}

// With the walk wide enough to reach multi-card target sets, none is
// offered with more targets than the payment's X, and X = 2 and 3 do get
// their pairs and triples: the "up to X" count follows the reveal.
func TestRevealXHoldsTheTargetCountToTheRevealedX(t *testing.T) {
	g, seat, moves, got := martyrBonesMovesBudget(t, 3, 60)
	widest := map[int]int{}
	for _, p := range got {
		if len(p.Targets) > p.XValue {
			t.Errorf("%d targets at X=%d", len(p.Targets), p.XValue)
		}
		widest[p.XValue] = max(widest[p.XValue], len(p.Targets))
	}
	if widest[1] != 1 || widest[2] != 2 || widest[3] != 3 {
		t.Errorf("widest target set per X = %v, want 1 / 2 / 3", widest)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// One black card pays one count only.
func TestRevealXWithAShortHand(t *testing.T) {
	_, _, _, got := martyrBonesMoves(t, 1)
	if len(got) == 0 {
		t.Fatal("no activation offered with one black card in hand")
	}
	for _, p := range got {
		if len(p.RevealIDs) != 1 || p.XValue != 1 {
			t.Errorf("payload %+v, want exactly the one-card payment", p)
		}
	}
}

// No black card: only X = 0 would pay, which does nothing, so nothing is
// offered.
func TestRevealXWithNoMatchingCardIsNotOffered(t *testing.T) {
	_, _, _, got := martyrBonesMoves(t, 0)
	if len(got) != 0 {
		t.Errorf("offered %d activations with no black card in hand: %+v", len(got), got)
	}
}

func cardInHand(t *testing.T, p *game.Player, id string) game.Card {
	t.Helper()
	want, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("bad id %q: %v", id, err)
	}
	for _, c := range p.Hand.Cards {
		if c.InstanceID == want {
			return c
		}
	}
	t.Fatalf("card %s is not in %s's hand", id, p.Name)
	return game.Card{}
}
