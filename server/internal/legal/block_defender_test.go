package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// block_defender_test.go — #1339: the enumerator and the wire verb
// agree about WHO may block an attacker (CR 802.4a), at a four-seat
// table with all three kinds of attack in the air at once.
//
// The enumerator always offered a seat only the attackers it defends
// (game.BlockOptionsLocked). The verb took any pairing whose blocker
// the caller controlled, so a human client could send a block no bot
// would ever be offered. This walks every (defending creature,
// attacker) pair through actions.Dispatch — the path a client's frame
// takes — and demands: offered ⇔ accepted.

func TestBlockMovesAndTheVerbAgreeOnTheDefendingPlayer(t *testing.T) {
	g := newTable(t)
	for _, p := range g.Seats {
		clearHand(p)
	}
	s0, s1, s2, s3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	creature := func(p *game.Player, name string) uuid.UUID {
		return battlefieldCard(g, p, game.Card{Name: name, TypeLine: "Creature — Test", Power: 3, Toughness: 3})
	}
	// Seat 3 controls the battle; seat 2 protects it. Seat 3 controls
	// the planeswalker.
	battle := battlefieldCard(g, s3, game.Card{
		Name: "Test Siege", TypeLine: "Battle — Siege",
		ProtectorPlayerID: s2.ID, Counters: map[string]int{game.CounterDefense: 5},
	})
	walker := battlefieldCard(g, s3, game.Card{
		Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	atPlayer := creature(s0, "At Seat 1")
	atBattle := creature(s0, "At The Battle")
	atWalker := creature(s0, "At The Walker")
	// defenderOf is what CR 802.4a says, written out by hand.
	defenderOf := map[uuid.UUID]uuid.UUID{atPlayer: s1.ID, atBattle: s2.ID, atWalker: s3.ID}
	blockers := map[uuid.UUID]uuid.UUID{
		creature(s1, "Seat 1 Blocker"): s1.ID,
		creature(s2, "Seat 2 Blocker"): s2.ID,
		creature(s3, "Seat 3 Blocker"): s3.ID,
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	for a, target := range map[uuid.UUID]uuid.UUID{atPlayer: s1.ID, atBattle: battle, atWalker: walker} {
		if err := g.DeclareAttacker(a, target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	offered := map[[2]uuid.UUID]bool{}
	for _, seat := range []*game.Player{s1, s2, s3} {
		for _, m := range legal.EnumerateFor(g, seat.ID) {
			if m.Kind != legal.KindBlock || m.Type != legal.TypeDeclareBlocker {
				continue
			}
			var p struct{ Blocker, Attacker string }
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("block move params: %v", err)
			}
			offered[[2]uuid.UUID{uuid.MustParse(p.Blocker), uuid.MustParse(p.Attacker)}] = true
		}
	}

	accepted := 0
	for b, owner := range blockers {
		for a, defender := range defenderOf {
			params, _ := json.Marshal(map[string]string{"blocker": b.String(), "attacker": a.String()})
			err := actions.Dispatch(g.Clone(), actions.Action{
				Type: actions.TypeDeclareBlocker, Player: owner, Caller: owner, Params: params,
			})
			want := owner == defender
			if (err == nil) != want {
				t.Errorf("blocker of %v on attacker defended by %v: verb err = %v, want legal = %v", owner, defender, err, want)
			}
			if offered[[2]uuid.UUID{b, a}] != want {
				t.Errorf("blocker of %v on attacker defended by %v: offered = %v, want %v", owner, defender, offered[[2]uuid.UUID{b, a}], want)
			}
			if want {
				accepted++
			}
		}
	}
	if accepted != 3 {
		t.Fatalf("the table has %d legal blocks, want 3 (one per defending seat) — the test proves nothing without them", accepted)
	}
}
