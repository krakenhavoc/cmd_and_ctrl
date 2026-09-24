package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestBlockMovesOfferAnAttackerWhoseWalkerLeft — #1364, CR 506.4c. The
// walker dies before blocks; its former controller is offered the
// block and the verb accepts it, and the other seat is offered nothing
// and refused.
func TestBlockMovesOfferAnAttackerWhoseWalkerLeft(t *testing.T) {
	g := newTable(t)
	for _, p := range g.Seats {
		clearHand(p)
	}
	s0, s1, s2 := g.Seats[0], g.Seats[1], g.Seats[2]
	walker := battlefieldCard(g, s2, game.Card{
		Name: "Doomed Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	attacker := battlefieldCard(g, s0, game.Card{Name: "Attacker", TypeLine: "Creature — Test", Power: 3, Toughness: 3})
	guard := battlefieldCard(g, s2, game.Card{Name: "Guard", TypeLine: "Creature — Test", Power: 1, Toughness: 4})
	bystander := battlefieldCard(g, s1, game.Card{Name: "Bystander", TypeLine: "Creature — Test", Power: 1, Toughness: 4})

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(walker) })
	if err != nil {
		t.Fatalf("destroy the walker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	wantBlockOffers(t, g, attacker, []blockOfferCase{{s2, guard, true}, {s1, bystander, false}})
}

// TestBlockMovesOfferAnAttackerWhoseWalkerWasStolen — #1376, CR 506.4c.
// The walker is stolen by the BYSTANDER's seat after the declaration:
// it leaves combat, the attacker attacks nothing, and the walker's
// controller before the removal (CR 802.2a) is still the one offered
// the block — not its new controller, whom the live read named.
func TestBlockMovesOfferAnAttackerWhoseWalkerWasStolen(t *testing.T) {
	g := newTable(t)
	for _, p := range g.Seats {
		clearHand(p)
	}
	s0, s1, s2 := g.Seats[0], g.Seats[1], g.Seats[2]
	walker := battlefieldCard(g, s2, game.Card{
		Name: "Stolen Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	attacker := battlefieldCard(g, s0, game.Card{Name: "Attacker", TypeLine: "Creature — Test", Power: 3, Toughness: 3})
	guard := battlefieldCard(g, s2, game.Card{Name: "Guard", TypeLine: "Creature — Test", Power: 1, Toughness: 4})
	bystander := battlefieldCard(g, s1, game.Card{Name: "Bystander", TypeLine: "Creature — Test", Power: 1, Toughness: 4})

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), walker, s1.ID, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatalf("GainControlForEffect refused")
		}
		g.RecomputeLayersIfStaleLocked()
	})

	wantBlockOffers(t, g, attacker, []blockOfferCase{{s2, guard, true}, {s1, bystander, false}})
}

type blockOfferCase struct {
	seat    *game.Player
	blocker uuid.UUID
	want    bool
}

// wantBlockOffers checks, per case, that the enumerator offers the
// block exactly when actions.Dispatch accepts it.
func wantBlockOffers(t *testing.T, g *game.Game, attacker uuid.UUID, cases []blockOfferCase) {
	t.Helper()
	offered := func(seat *game.Player, blocker uuid.UUID) bool {
		for _, m := range legal.EnumerateFor(g, seat.ID) {
			if m.Kind != legal.KindBlock || m.Type != legal.TypeDeclareBlocker {
				continue
			}
			var p struct{ Blocker, Attacker string }
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("block move params: %v", err)
			}
			if p.Blocker == blocker.String() && p.Attacker == attacker.String() {
				return true
			}
		}
		return false
	}
	for _, tc := range cases {
		if got := offered(tc.seat, tc.blocker); got != tc.want {
			t.Errorf("%s offered the block = %v, want %v", tc.seat.Name, got, tc.want)
		}
		params, _ := json.Marshal(map[string]string{"blocker": tc.blocker.String(), "attacker": attacker.String()})
		err := actions.Dispatch(g.Clone(), actions.Action{
			Type: actions.TypeDeclareBlocker, Player: tc.seat.ID, Caller: tc.seat.ID, Params: params,
		})
		if (err == nil) != tc.want {
			t.Errorf("%s blocks: err = %v, want legal = %v", tc.seat.Name, err, tc.want)
		}
	}
}
