package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_limit_view_test.go — #1533, ADR 0045 Decision 46: the wire
// half of the attack-with-all picker's cap. `turn.attack_targets[]`
// carries `attack_limit`, the room a CR 508.1c count limit leaves at
// that target, so the client caps its selection without re-deriving
// who the limit protects or how many are already attacking.

const attackLimitViewOracle = "attack-limit-view-test"

// stubViewAttackLimits answers `limits` for attackLimitViewOracle and
// nothing else, and restores the hook afterwards.
func stubViewAttackLimits(t *testing.T, limits ...game.AttackLimit) {
	t.Helper()
	prev := game.CatalogAttackLimits
	t.Cleanup(func() { game.CatalogAttackLimits = prev })
	game.CatalogAttackLimits = func(key string) []game.AttackLimit {
		if key == attackLimitViewOracle {
			return limits
		}
		return nil
	}
}

// threeSeatsInDeclareAttackers is a started three-player game advanced
// to the active player's declare_attackers step.
func threeSeatsInDeclareAttackers(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for i := 0; i < 20 && g.Turn.Step != game.StepDeclareAttackers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatal("never reached declare_attackers")
	}
	return g
}

// pushLimitPermanent puts a permanent carrying the stubbed limits onto
// the battlefield under `controller`.
func pushLimitPermanent(g *game.Game, controller uuid.UUID) {
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Limit Source", TypeLine: "Artifact",
		OracleID: attackLimitViewOracle, Owner: controller, Controller: controller,
	})
}

// pushAttacker puts a creature under `controller` onto the battlefield,
// attacking `target` when it is not uuid.Nil.
func pushAttacker(g *game.Game, controller, target uuid.UUID) {
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: controller, Controller: controller,
		AttackingTarget: target,
	})
}

// attackLimitsByTarget reads every attack_targets row's attack_limit:
// -1 for an absent field.
func attackLimitsByTarget(v GameView) map[string]int {
	out := map[string]int{}
	for _, row := range v.Turn.AttackTargets {
		if row.AttackLimit == nil {
			out[row.ID] = -1
			continue
		}
		out[row.ID] = *row.AttackLimit
	}
	return out
}

// TestAttackTargetsCarryTheAttackLimitRoom — Crawlspace (two may
// attack its controller) with one creature already attacking that
// seat. Its row says one more; the other opponent's row and the
// Crawlspace player's planeswalker carry no field, because "you" is
// the player alone.
func TestAttackTargetsCarryTheAttackLimitRoom(t *testing.T) {
	stubViewAttackLimits(t, game.AttackLimit{Scope: game.AttackLimitAttackingYou, Max: 2})
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	crawl := g.Seats[(g.Turn.ActiveSeat+1)%3]
	other := g.Seats[(g.Turn.ActiveSeat+2)%3]
	pushLimitPermanent(g, crawl.ID)
	walker := seatWalker(g, crawl.ID, 4)
	pushAttacker(g, active.ID, crawl.ID)
	pushAttacker(g, active.ID, uuid.Nil)

	got := attackLimitsByTarget(ViewOfGame(g))
	if got[crawl.ID.String()] != 1 {
		t.Errorf("the Crawlspace seat's attack_limit = %d, want 1", got[crawl.ID.String()])
	}
	if got[other.ID.String()] != -1 {
		t.Errorf("an unprotected seat carries attack_limit %d", got[other.ID.String()])
	}
	if got[walker.String()] != -1 {
		t.Errorf("the Crawlspace player's planeswalker carries attack_limit %d", got[walker.String()])
	}
}

// TestAttackTargetsCarryAZeroAttackLimit — Silent Arbiter with its one
// attacker already declared. 0 is a real answer (the limit is used up),
// so it is on the wire at every target rather than dropped by
// omitempty and read as "unlimited".
func TestAttackTargetsCarryAZeroAttackLimit(t *testing.T) {
	stubViewAttackLimits(t, game.AttackLimit{Scope: game.AttackLimitEachCombat, Max: 1})
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%3]
	pushLimitPermanent(g, defender.ID)
	pushAttacker(g, active.ID, defender.ID)
	pushAttacker(g, active.ID, uuid.Nil)

	v := ViewOfGame(g)
	for id, n := range attackLimitsByTarget(v) {
		if n != 0 {
			t.Errorf("target %s: attack_limit = %d, want 0", id, n)
		}
	}
	raw, err := json.Marshal(v.Turn)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"attack_limit":0`) {
		t.Errorf("a used-up limit is not on the wire as 0: %s", raw)
	}
}

// TestAttackTargetsOmitTheAttackLimitWithoutOne — no limit on the
// battlefield: the field is absent from every row, which is the whole
// wire cost at nearly every table.
func TestAttackTargetsOmitTheAttackLimitWithoutOne(t *testing.T) {
	stubViewAttackLimits(t)
	g := threeSeatsInDeclareAttackers(t)
	pushAttacker(g, g.Seats[g.Turn.ActiveSeat].ID, uuid.Nil)

	v := ViewOfGame(g)
	if len(v.Turn.AttackTargets) == 0 {
		t.Fatal("no attack targets published")
	}
	raw, err := json.Marshal(v.Turn)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "attack_limit") {
		t.Errorf("attack_limit on the wire with no limit in play: %s", raw)
	}
}
