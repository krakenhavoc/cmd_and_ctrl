package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// hybrid_phyrexian_test.go — the enumerator half of #787. CR 107.4's
// hybrid Phyrexian symbols used to make ParseCost fail, and the #289
// guard in castMovesForCard then dropped the card: a compleated
// planeswalker could not be cast at all, from the board or by a bot.
//
// The cast the enumerator offers is the MANA payment — the life half
// (CR 107.4f) is an announce-time claim the caster makes, not a
// second move. That keeps #695's complaint from spreading: nothing
// here advertises a life payment a life total could not cover,
// because nothing here advertises a life payment.

// TestHybridPhyrexianCastIsEnumerated — with mana for either half of
// the symbol, the cast is offered.
func TestHybridPhyrexianCastIsEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	// Ajani, Sleeper Agent's printed cost, on a sorcery so nothing
	// but the cost is under test.
	sage := handCard(active, game.Card{
		Name:     "Compleated Sage",
		TypeLine: "Sorcery",
		ManaCost: "{1}{G}{G/W/P}{W}",
		Layout:   "normal",
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Plains", "Plains"))
	battlefieldCard(g, active, basic("Plains", "Plains"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	seen := false
	for _, m := range moves {
		if m.Source == sage {
			seen = true
		}
	}
	if !seen {
		t.Errorf("a card printing {G/W/P} should be castable off four lands: %v", labels(moves))
	}
}

// TestHybridPhyrexianCastNeedsTheManaForTheSymbol — the offer is still
// affordability-gated: three lands cannot pay a four-symbol cost, and
// the enumerator does not quietly treat the Phyrexian symbol as free
// just because it has a life alternative the move does not claim.
func TestHybridPhyrexianCastNeedsTheManaForTheSymbol(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sage := handCard(active, game.Card{
		Name:     "Compleated Sage",
		TypeLine: "Sorcery",
		ManaCost: "{1}{G}{G/W/P}{W}",
		Layout:   "normal",
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Plains", "Plains"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	for _, m := range moves {
		if m.Source == sage {
			t.Errorf("offered a cast three lands cannot pay: %q", m.Label)
		}
	}
}

// --- #917: the ACTIVATION half ------------------------------------
//
// An activated ability's mana component has an announce for the life
// half now (ActivateAbilityParams.PhyrexianLife, CR 602.2b), so the
// enumerator solves for the pair: the X and how many symbols the
// life buys. Unlike the cast above it DOES offer the life payment,
// and the reason the two differ is the reason #695 gave — an offer
// is safe exactly when it is bounded by the life total, which this
// one is and a cast's unbounded one would not have been.

// phyrexianLifeOf pulls phyrexian_life out of an activate_ability
// move's params.
func phyrexianLifeOf(t *testing.T, m legal.Move) int {
	t.Helper()
	var p struct {
		PhyrexianLife int `json:"phyrexian_life"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.PhyrexianLife
}

// phyrexianPod seats Birthing Pod's cost shape on the battlefield:
// one ability costing "{1}{G/P}" and nothing else.
func phyrexianPod(g *game.Game, p *game.Player) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name:     "Phyrexian Pod",
		TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{1}{G/P}: Mark",
			Cost:  game.AbilityCost{Mana: "{1}{G/P}"},
			Effect: func(_ *game.Game, _ *game.StackItem) error {
				return nil
			},
		}},
	})
}

// TestActivationPrefersManaOverLife — with a green source on the
// board the enumerator pays the symbol with mana and claims nothing.
// A life total is a resource, and the list must not spend it to save
// mana the board could have produced.
func TestActivationPrefersManaOverLife(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pod := phyrexianPod(g, active)
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	got := activationsOf(moves, pod)
	if len(got) != 1 {
		t.Fatalf("activations of the Pod = %d, want 1: %v", len(got), labels(moves))
	}
	if n := phyrexianLifeOf(t, got[0]); n != 0 {
		t.Errorf("phyrexian_life = %d, want 0 — two Forests pay the {G/P} with mana", n)
	}
	dispatchAll(t, g, active.ID, moves)
}

// TestActivationOffersTheLifeOptionWhenManaIsShort — #917's gap: with
// no green source the {G/P} is payable only with 2 life, and before
// this the ability was simply not offered. The move carries the
// claim and the price.
func TestActivationOffersTheLifeOptionWhenManaIsShort(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pod := phyrexianPod(g, active)
	// One blue source: it pays the {1} and nothing else.
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	got := activationsOf(moves, pod)
	if len(got) != 1 {
		t.Fatalf("activations of the Pod = %d, want 1: %v", len(got), labels(moves))
	}
	if n := phyrexianLifeOf(t, got[0]); n != 1 {
		t.Errorf("phyrexian_life = %d, want 1 — the {G/P} has to be bought with life", n)
	}
	// #74: the price rides the Move, so a policy can see what the
	// activation costs in life.
	if got[0].Cost == nil || got[0].Cost.Life != game.PhyrexianLifePerSymbol {
		t.Errorf("move cost = %+v, want Life = %d", got[0].Cost, game.PhyrexianLifePerSymbol)
	}
	// #544: the engine accepts exactly what the list offered.
	dispatchAll(t, g, active.ID, moves)
}

// TestActivationDoesNotOfferLifeItCannotPay — CR 119.4 bounds the
// offer, which is the whole reason offering one is safe (#695). At 2
// life the claim is still legal (paying down to exactly 0 is); at 1
// it is not, and the ability is not offered at all.
func TestActivationDoesNotOfferLifeItCannotPay(t *testing.T) {
	for _, tc := range []struct {
		life  int
		offer bool
	}{
		{life: game.PhyrexianLifePerSymbol, offer: true},
		{life: game.PhyrexianLifePerSymbol - 1, offer: false},
	} {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		clearHand(active)
		pod := phyrexianPod(g, active)
		battlefieldCard(g, active, basic("Island", "Island"))
		advanceTo(t, g, game.StepPrecombatMain)
		active.Life = tc.life

		moves := legal.EnumerateFor(g, active.ID)
		got := activationsOf(moves, pod)
		if tc.offer && len(got) != 1 {
			t.Errorf("at %d life: %d activations, want 1", tc.life, len(got))
		}
		if !tc.offer && len(got) != 0 {
			t.Errorf("at %d life: offered an activation CR 119.4 refuses: %q", tc.life, got[0].Label)
		}
		if tc.offer {
			dispatchAll(t, g, active.ID, moves)
		}
	}
}
