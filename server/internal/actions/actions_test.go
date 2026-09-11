package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func newGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// S13: close the mulligan window so the cursor lands on a
	// priority-granting step. Tests in this package exercise priority/
	// step gating against a real *game.Game; without closing mulligans
	// the cursor sits at Untap with PriorityHolder=NoPriority and
	// most pass_priority dispatches would fail with ErrNoPriority.
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand %s: %v", p.Name, err)
		}
	}
	return g
}

func params(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	return raw
}

func TestDispatchDrawCard(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]

	a, err := Decode(string(TypeDrawCard), p.ID.String(), nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	// Opening hand is 7 (dealt by Start as of S08); +1 from this draw.
	if p.Hand.Size() != 8 {
		t.Errorf("hand size: got %d, want 8", p.Hand.Size())
	}
}

func TestDispatchPlayCard(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	card, _ := p.Hand.Top()

	a, _ := Decode(
		string(TypePlayCard),
		p.ID.String(),
		params(t, map[string]string{"instance_id": card.InstanceID.String()}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Error("card not on battlefield after play")
	}
}

func TestDispatchMoveCard(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	a, _ := Decode(
		string(TypeMoveCard),
		"",
		params(t, map[string]any{
			"src":         map[string]string{"kind": "battlefield"},
			"dst":         map[string]string{"kind": "exile"},
			"instance_id": card.InstanceID.String(),
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("card not in exile after move_card")
	}
}

func TestDispatchTapAndUntap(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	tapAction, _ := Decode(
		string(TypeTap),
		"",
		params(t, map[string]string{"instance_id": card.InstanceID.String()}),
	)
	if err := Dispatch(g, tapAction); err != nil {
		t.Fatalf("tap Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && !c.Tapped {
			t.Error("card should be tapped after tap action")
		}
	}

	untapAction, _ := Decode(
		string(TypeUntap),
		"",
		params(t, map[string]string{"instance_id": card.InstanceID.String()}),
	)
	if err := Dispatch(g, untapAction); err != nil {
		t.Fatalf("untap Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Tapped {
			t.Error("card should be untapped after untap action")
		}
	}
}

func TestDispatchUntapAllRequiresPlayer(t *testing.T) {
	g := newGame(t)
	a, _ := Decode(string(TypeUntapAll), "", nil)
	if err := Dispatch(g, a); !errors.Is(err, ErrInvalidPlayer) {
		t.Errorf("untap_all without player: got %v, want ErrInvalidPlayer", err)
	}
}

func TestDispatchPassPriorityRotates(t *testing.T) {
	// 2-player game: one PassPriority rotates priority to seat 1
	// without advancing the step (the next would wrap and advance).
	g := newGame(t)
	beforeStep := g.Turn.Step
	beforeHolder := g.Turn.PriorityHolder
	a, _ := Decode(string(TypePassPriority), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Turn.Step != beforeStep {
		t.Errorf("step changed unexpectedly: %q -> %q", beforeStep, g.Turn.Step)
	}
	if g.Turn.PriorityHolder == beforeHolder {
		t.Errorf("priority did not rotate (still %d)", g.Turn.PriorityHolder)
	}
}

func TestDispatchPassTurnWrapsSeat(t *testing.T) {
	g := newGame(t)
	a, _ := Decode(string(TypePassTurn), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("seat after pass_turn: got %d, want 1", g.Turn.ActiveSeat)
	}
}

func TestDispatchPassPriorityRejectsNonHolder(t *testing.T) {
	g := newGame(t)
	holder := g.Seats[g.Turn.PriorityHolder].ID
	other := g.Seats[(g.Turn.PriorityHolder+1)%len(g.Seats)].ID

	a, _ := Decode(string(TypePassPriority), "", nil)
	a.Caller = other
	if err := Dispatch(g, a); !errors.Is(err, ErrNotPriorityHolder) {
		t.Errorf("non-holder dispatch: got %v, want ErrNotPriorityHolder", err)
	}
	// State must not have moved.
	if g.Seats[g.Turn.PriorityHolder].ID != holder {
		t.Errorf("priority holder changed despite rejected dispatch")
	}

	// The legitimate holder still succeeds.
	a.Caller = holder
	if err := Dispatch(g, a); err != nil {
		t.Errorf("holder dispatch: got %v, want nil", err)
	}
}

func TestDispatchPassTurnRejectsNonActivePlayer(t *testing.T) {
	g := newGame(t)
	active := g.Seats[g.Turn.ActiveSeat].ID
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID

	a, _ := Decode(string(TypePassTurn), "", nil)
	a.Caller = other
	if err := Dispatch(g, a); !errors.Is(err, ErrNotActivePlayer) {
		t.Errorf("non-active dispatch: got %v, want ErrNotActivePlayer", err)
	}
	if g.Seats[g.Turn.ActiveSeat].ID != active {
		t.Errorf("active seat changed despite rejected dispatch")
	}

	// The active player still succeeds.
	a.Caller = active
	if err := Dispatch(g, a); err != nil {
		t.Errorf("active-player dispatch: got %v, want nil", err)
	}
}

func TestDispatchRejectsCrossSeatPlayerScopedAction(t *testing.T) {
	g := newGame(t)
	caller := g.Seats[0].ID
	target := g.Seats[1].ID

	// change_life on someone else's seat is the prototypical case.
	a, _ := Decode(string(TypeChangeLife), target.String(), params(t, struct {
		Delta int `json:"delta"`
	}{Delta: -40}))
	a.Caller = caller
	if err := Dispatch(g, a); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("cross-seat change_life: got %v, want ErrPlayerCallerMismatch", err)
	}
	if g.Seats[1].Life != 40 {
		t.Errorf("target life changed despite rejected dispatch: %d", g.Seats[1].Life)
	}

	// Self-targeting still works.
	a, _ = Decode(string(TypeChangeLife), caller.String(), params(t, struct {
		Delta int `json:"delta"`
	}{Delta: -3}))
	a.Caller = caller
	if err := Dispatch(g, a); err != nil {
		t.Errorf("self-targeted change_life: got %v, want nil", err)
	}
	if g.Seats[0].Life != 37 {
		t.Errorf("self life: got %d, want 37", g.Seats[0].Life)
	}
}

// pushCreature drops a 2/2 test creature onto the battlefield and
// returns its instance ID as a string. Combat dispatch tests need
// real creatures (the game's IsCreature check rejects placeholder
// cards from the demo deck).
func pushCreature(t *testing.T, g *game.Game, p *game.Player) string {
	t.Helper()
	c := game.NewCard("Test Creature", p.ID)
	c.TypeLine = "Creature — Test"
	c.Power = 2
	c.Toughness = 2
	g.Battlefield.PushTop(c)
	return c.InstanceID.String()
}

// advanceTo walks the game's step cursor forward to the named step.
func advanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for g.Turn.Step != step {
		before := g.Turn
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn == before {
			t.Fatalf("advanceTo loop didn't progress past %v", before)
		}
	}
}

func TestDispatchDeclareAttacker(t *testing.T) {
	g := newGame(t)
	attacker := pushCreature(t, g, g.Seats[0])
	advanceTo(t, g, game.StepDeclareAttackers)
	target := g.Seats[1].ID.String()

	a, _ := Decode(string(TypeDeclareAttacker), "", params(t, map[string]string{
		"attacker": attacker,
		"target":   target,
	}))
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == attacker && c.AttackingTarget.String() != target {
			t.Errorf("AttackingTarget: got %v, want %v", c.AttackingTarget, target)
		}
	}
}

func TestDispatchDeclareBlocker(t *testing.T) {
	g := newGame(t)
	attacker := pushCreature(t, g, g.Seats[0])
	blocker := pushCreature(t, g, g.Seats[1])
	advanceTo(t, g, game.StepDeclareAttackers)
	_ = Dispatch(g, mustAction(t, TypeDeclareAttacker, params(t, map[string]string{
		"attacker": attacker,
		"target":   g.Seats[1].ID.String(),
	})))
	advanceTo(t, g, game.StepDeclareBlockers)

	a, _ := Decode(string(TypeDeclareBlocker), "", params(t, map[string]string{
		"blocker":  blocker,
		"attacker": attacker,
	}))
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == blocker && c.BlockingTarget.String() != attacker {
			t.Errorf("BlockingTarget: got %v, want %v", c.BlockingTarget, attacker)
		}
	}
}

func TestDispatchClearCombat(t *testing.T) {
	g := newGame(t)
	attacker := pushCreature(t, g, g.Seats[0])
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := Dispatch(g, mustAction(t, TypeDeclareAttacker, params(t, map[string]string{
		"attacker": attacker,
		"target":   g.Seats[1].ID.String(),
	}))); err != nil {
		t.Fatalf("Dispatch declare: %v", err)
	}

	clear, _ := Decode(string(TypeClearCombat), "", nil)
	if err := Dispatch(g, clear); err != nil {
		t.Fatalf("Dispatch clear: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			t.Errorf("card %v still attacking after clear_combat", c.InstanceID)
		}
	}
}

// mustAction is a thin Decode wrapper that fails the test on decode
// error. Used so combat dispatch tests can chain multiple actions
// without local error boilerplate.
func mustAction(t *testing.T, ty Type, p []byte) Action {
	t.Helper()
	a, err := Decode(string(ty), "", p)
	if err != nil {
		t.Fatalf("Decode %s: %v", ty, err)
	}
	return a
}

func TestDispatchKeepHand(t *testing.T) {
	g := newGame(t)
	caller := g.Seats[0].ID
	a, _ := Decode(string(TypeKeepHand), caller.String(), nil)
	a.Caller = caller
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !g.Seats[0].HandKept {
		t.Error("seat 0 HandKept not set")
	}
}

func TestDispatchKeepHandRejectsCrossSeat(t *testing.T) {
	g := newGame(t)
	caller := g.Seats[0].ID
	target := g.Seats[1].ID
	a, _ := Decode(string(TypeKeepHand), target.String(), nil)
	a.Caller = caller
	if err := Dispatch(g, a); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("cross-seat keep_hand: got %v, want ErrPlayerCallerMismatch", err)
	}
}

func TestDispatchConcede(t *testing.T) {
	g := newGame(t)
	caller := g.Seats[0].ID
	a, _ := Decode(string(TypeConcede), caller.String(), nil)
	a.Caller = caller
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !g.Seats[0].Eliminated {
		t.Error("seat 0 should be eliminated after concede")
	}
}

func TestDispatchConcedeRequiresPlayer(t *testing.T) {
	g := newGame(t)
	a, _ := Decode(string(TypeConcede), "", nil)
	if err := Dispatch(g, a); !errors.Is(err, ErrInvalidPlayer) {
		t.Errorf("concede without player: got %v, want ErrInvalidPlayer", err)
	}
}

func TestDispatchConcedeRejectsCrossSeat(t *testing.T) {
	g := newGame(t)
	caller := g.Seats[0].ID
	target := g.Seats[1].ID
	a, _ := Decode(string(TypeConcede), target.String(), nil)
	a.Caller = caller
	if err := Dispatch(g, a); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("cross-seat concede: got %v, want ErrPlayerCallerMismatch", err)
	}
	if g.Seats[1].Eliminated {
		t.Error("target wrongly marked eliminated despite rejected dispatch")
	}
}

func TestDispatchPassPriorityAllowsAdminCaller(t *testing.T) {
	// Admin / spectator (Caller == uuid.Nil) bypasses the gate so a
	// trusted moderator can advance the game on a player's behalf.
	g := newGame(t)
	a, _ := Decode(string(TypePassPriority), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Errorf("admin dispatch: got %v, want nil", err)
	}
}

func TestDispatchMulligan(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	for range 7 {
		_ = g.DrawCard(p.ID)
	}
	a, _ := Decode(
		string(TypeMulligan),
		p.ID.String(),
		params(t, map[string]int{"hand_size": 6}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p.Hand.Size() != 6 {
		t.Errorf("hand size: got %d, want 6", p.Hand.Size())
	}
}

func TestDispatchShuffleLibrary(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	before := append([]game.Card(nil), p.Library.Cards...)
	a, _ := Decode(string(TypeShuffleLibrary), p.ID.String(), nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	diff := false
	for i := range before {
		if before[i].InstanceID != p.Library.Cards[i].InstanceID {
			diff = true
			break
		}
	}
	if !diff {
		t.Error("library unchanged after shuffle")
	}
}

func TestDispatchChangeLife(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	a, _ := Decode(
		string(TypeChangeLife),
		p.ID.String(),
		params(t, map[string]int{"delta": -3}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p.Life != 37 {
		t.Errorf("life: got %d, want 37", p.Life)
	}
}

func TestDispatchAddCounter(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	a, _ := Decode(
		string(TypeAddCounter),
		"",
		params(t, map[string]any{
			"instance_id": card.InstanceID.String(),
			"name":        "+1/+1",
			"delta":       2,
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Counters["+1/+1"] != 2 {
			t.Errorf("counter: got %d, want 2", c.Counters["+1/+1"])
		}
	}
}

// TestDispatchSetCommanderDamage — `from` is a commander CARD
// instance ID since S25 (#77), not the opposing player's ID. The
// commander is looked up wherever it lives, which for a freshly
// started game is the command zone.
func TestDispatchSetCommanderDamage(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	cmdr := p0.Command.Cards[0].InstanceID
	a, _ := Decode(
		string(TypeSetCommanderDamage),
		"",
		params(t, map[string]any{
			"from":   cmdr.String(),
			"to":     p1.ID.String(),
			"amount": 12,
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p1.CommanderDamage[cmdr] != 12 {
		t.Errorf("cmdr damage: got %d, want 12", p1.CommanderDamage[cmdr])
	}
}

// TestDispatchSetCommanderDamageRejectsAPlayerID pins the rekey: the
// old call shape (a player ID in `from`) must now fail loudly rather
// than write a key the client can never render.
func TestDispatchSetCommanderDamageRejectsAPlayerID(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	a, _ := Decode(
		string(TypeSetCommanderDamage),
		"",
		params(t, map[string]any{
			"from":   p0.ID.String(),
			"to":     p1.ID.String(),
			"amount": 12,
		}),
	)
	if err := Dispatch(g, a); err == nil {
		t.Fatal("Dispatch with a player ID in `from`: got nil, want an error")
	}
	if len(p1.CommanderDamage) != 0 {
		t.Errorf("CommanderDamage was written despite the rejection: %v", p1.CommanderDamage)
	}
}

func TestDispatchSetBattlefieldPosition(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	a, _ := Decode(
		string(TypeSetBattlefieldPosition),
		"",
		params(t, map[string]any{
			"instance_id": card.InstanceID.String(),
			"x":           0.4,
			"y":           0.6,
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID {
			if c.BattleX != 0.4 || c.BattleY != 0.6 {
				t.Errorf("position: got (%v, %v), want (0.4, 0.6)", c.BattleX, c.BattleY)
			}
		}
	}
}

func TestDispatchSetMonarch(t *testing.T) {
	g := newGame(t)
	p0 := g.Seats[0]
	a, _ := Decode(string(TypeSetMonarch), p0.ID.String(), nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Monarch != p0.ID {
		t.Errorf("monarch: got %v, want %v", g.Monarch, p0.ID)
	}
	// Empty player clears.
	clear, _ := Decode(string(TypeSetMonarch), "", nil)
	if err := Dispatch(g, clear); err != nil {
		t.Fatalf("Dispatch clear: %v", err)
	}
	if g.Monarch != uuid.Nil {
		t.Errorf("monarch cleared: got %v, want nil", g.Monarch)
	}
}

func TestDispatchSetInitiative(t *testing.T) {
	g := newGame(t)
	p0 := g.Seats[0]
	a, _ := Decode(string(TypeSetInitiative), p0.ID.String(), nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Initiative != p0.ID {
		t.Errorf("initiative: got %v, want %v", g.Initiative, p0.ID)
	}
}

func TestDispatchSetGoaded(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Atog", TypeLine: "Creature — Atog",
		Owner: p0.ID, Controller: p0.ID,
	})
	c := g.Battlefield.Cards[0]
	a, _ := Decode(string(TypeSetGoaded), "", params(t, map[string]string{
		"instance_id": c.InstanceID.String(),
		"by":          p1.ID.String(),
	}))
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Battlefield.Cards[0].GoadedBy != p1.ID {
		t.Errorf("goaded_by: got %v, want %v", g.Battlefield.Cards[0].GoadedBy, p1.ID)
	}
}

func TestDispatchSetPoison(t *testing.T) {
	g := newGame(t)
	p0 := g.Seats[0]
	a, _ := Decode(string(TypeSetPoison), p0.ID.String(), params(t, map[string]int{"amount": 5}))
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p0.Poison != 5 {
		t.Errorf("poison: got %d, want 5", p0.Poison)
	}
}

func TestDispatchSetEnergyRejectsCrossSeat(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	a, _ := Decode(string(TypeSetEnergy), p1.ID.String(), params(t, map[string]int{"amount": 4}))
	a.Caller = p0.ID
	err := Dispatch(g, a)
	if !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("cross-seat energy: got %v, want ErrPlayerCallerMismatch", err)
	}
}

func TestDispatchSetPromise(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	a, _ := Decode(string(TypeSetPromise), "", params(t, map[string]any{
		"from":  p0.ID.String(),
		"to":    p1.ID.String(),
		"count": 2,
	}))
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Promises[game.PromiseKey{From: p0.ID, To: p1.ID}] != 2 {
		t.Errorf("promise: got %d, want 2", g.Promises[game.PromiseKey{From: p0.ID, To: p1.ID}])
	}
}

func TestDispatchVoteFlow(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]

	start, _ := Decode(string(TypeStartVote), p0.ID.String(), params(t, map[string]any{
		"topic":   "monarch?",
		"options": []string{"alice", "bob"},
	}))
	if err := Dispatch(g, start); err != nil {
		t.Fatalf("Dispatch start_vote: %v", err)
	}
	if g.Vote == nil {
		t.Fatal("vote did not open")
	}

	cast0, _ := Decode(string(TypeCastVote), p0.ID.String(), params(t, map[string]int{"option": 0}))
	cast1, _ := Decode(string(TypeCastVote), p1.ID.String(), params(t, map[string]int{"option": 1}))
	if err := Dispatch(g, cast0); err != nil {
		t.Fatalf("Dispatch cast_vote p0: %v", err)
	}
	if err := Dispatch(g, cast1); err != nil {
		t.Fatalf("Dispatch cast_vote p1: %v", err)
	}
	if g.Vote.Ballots[p0.ID] != 0 || g.Vote.Ballots[p1.ID] != 1 {
		t.Errorf("ballots: %v", g.Vote.Ballots)
	}

	end, _ := Decode(string(TypeEndVote), "", nil)
	if err := Dispatch(g, end); err != nil {
		t.Fatalf("Dispatch end_vote: %v", err)
	}
	if g.Vote != nil {
		t.Error("vote did not close")
	}
}

func TestDispatchCastVoteRejectsCrossSeat(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	_, _ = g.StartVote(p0.ID, "x", []string{"a", "b"})
	a, _ := Decode(string(TypeCastVote), p1.ID.String(), params(t, map[string]int{"option": 0}))
	a.Caller = p0.ID
	if err := Dispatch(g, a); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("cross-seat cast_vote: got %v, want ErrPlayerCallerMismatch", err)
	}
}

func TestDispatchUnknownType(t *testing.T) {
	g := newGame(t)
	a, _ := Decode("explode", "", nil)
	err := Dispatch(g, a)
	if !errors.Is(err, ErrUnknownType) {
		t.Errorf("unknown type: got %v, want ErrUnknownType", err)
	}
}

func TestDecodeMalformedPlayer(t *testing.T) {
	_, err := Decode(string(TypeDrawCard), "not-a-uuid", nil)
	if !errors.Is(err, ErrInvalidPlayer) {
		t.Errorf("bad player uuid: got %v, want ErrInvalidPlayer", err)
	}
}

func TestDispatchMalformedParams(t *testing.T) {
	g := newGame(t)
	a := Action{
		Type:   TypePlayCard,
		Player: g.Seats[0].ID,
		Params: json.RawMessage(`{ not json`),
	}
	if err := Dispatch(g, a); err == nil {
		t.Error("expected error on malformed params")
	}
}

// TestCardActionRejectsWrongController is the S08.5 wave-1 controller
// gate. Player B issuing a card-instance action against Player A's
// creature must surface ErrCardCallerMismatch — not silently mutate
// somebody else's cards (which the loose pre-S08.5 dispatch did).
func TestCardActionRejectsWrongController(t *testing.T) {
	cases := []struct {
		name    string
		buildAt game.Step
		ty      Type
		params  func(cardID string, p0, p1 *game.Player) json.RawMessage
	}{
		{
			name: "tap",
			ty:   TypeTap,
			params: func(cardID string, _, _ *game.Player) json.RawMessage {
				return mustJSON(map[string]string{"instance_id": cardID})
			},
		},
		{
			name: "move_card",
			ty:   TypeMoveCard,
			params: func(cardID string, _, _ *game.Player) json.RawMessage {
				return mustJSON(map[string]any{
					"src":         map[string]string{"kind": "battlefield"},
					"dst":         map[string]string{"kind": "exile"},
					"instance_id": cardID,
				})
			},
		},
		{
			name: "add_counter",
			ty:   TypeAddCounter,
			params: func(cardID string, _, _ *game.Player) json.RawMessage {
				return mustJSON(map[string]any{
					"instance_id": cardID,
					"name":        "+1/+1",
					"delta":       1,
				})
			},
		},
		{
			name: "set_battlefield_position",
			ty:   TypeSetBattlefieldPosition,
			params: func(cardID string, _, _ *game.Player) json.RawMessage {
				return mustJSON(map[string]any{
					"instance_id": cardID,
					"x":           0.5,
					"y":           0.5,
				})
			},
		},
		{
			name:    "declare_attacker",
			buildAt: game.StepDeclareAttackers,
			ty:      TypeDeclareAttacker,
			params: func(cardID string, _, p1 *game.Player) json.RawMessage {
				return mustJSON(map[string]string{
					"attacker": cardID,
					"target":   p1.ID.String(),
				})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newGame(t)
			p0, p1 := g.Seats[0], g.Seats[1]
			cardID := pushCreature(t, g, p0)
			if tc.buildAt != "" {
				advanceTo(t, g, tc.buildAt)
			}
			a, err := Decode(string(tc.ty), "", tc.params(cardID, p0, p1))
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			a.Caller = p1.ID // wrong controller
			if err := Dispatch(g, a); !errors.Is(err, game.ErrCardCallerMismatch) {
				t.Errorf("got %v, want ErrCardCallerMismatch", err)
			}
		})
	}
}

// TestAdminBypassesCardControllerGate confirms that uuid.Nil callers
// (admin / spectator sessions) skip the controller gate so a moderator
// can fix a wedged board state.
func TestAdminBypassesCardControllerGate(t *testing.T) {
	g := newGame(t)
	p0 := g.Seats[0]
	cardID := pushCreature(t, g, p0)
	a, _ := Decode(string(TypeTap), "", mustJSON(map[string]string{"instance_id": cardID}))
	// a.Caller stays uuid.Nil — admin connection.
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("admin tap: %v", err)
	}
}

// TestCardActionAllowsCorrectController ensures the gate isn't
// over-zealous: the actual controller can still act on their own card.
func TestCardActionAllowsCorrectController(t *testing.T) {
	g := newGame(t)
	p0 := g.Seats[0]
	cardID := pushCreature(t, g, p0)
	a, _ := Decode(string(TypeTap), "", mustJSON(map[string]string{"instance_id": cardID}))
	a.Caller = p0.ID
	if err := Dispatch(g, a); err != nil {
		t.Errorf("controller tap: %v", err)
	}
}

// TestCardActionUnknownInstancePassesThrough verifies that a card-not-
// found case is surfaced via the canonical ErrCardNotFound, not via
// ErrCardCallerMismatch — clients should be able to distinguish "I
// don't know about that card" from "you don't control it".
func TestCardActionUnknownInstancePassesThrough(t *testing.T) {
	g := newGame(t)
	bogus := uuid.New().String()
	a, _ := Decode(string(TypeTap), "", mustJSON(map[string]string{"instance_id": bogus}))
	a.Caller = g.Seats[0].ID
	err := Dispatch(g, a)
	if errors.Is(err, game.ErrCardCallerMismatch) {
		t.Errorf("got ErrCardCallerMismatch, want ErrCardNotFound for unknown card")
	}
	if !errors.Is(err, game.ErrCardNotFound) {
		t.Errorf("got %v, want ErrCardNotFound", err)
	}
}

// mustJSON marshals v or panics. Test-only convenience for the table-
// driven controller-gate cases above.
func mustJSON(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}
