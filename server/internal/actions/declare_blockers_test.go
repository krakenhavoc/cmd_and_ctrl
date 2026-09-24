package actions

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declare_blockers_test.go covers the set-shaped declare_blockers verb
// (#750) at the dispatch layer: shape validation, the batch cap, the
// controller gate, and that it is all-or-nothing.
//
// Unlike declare_attackers, this verb exists for a RULES reason: a
// block COUNT is a property of the whole declaration, so a
// two-creature menace block cannot be sent as two declare_blocker
// actions — the first would be refused.

type blockerEntry struct {
	Blocker  string `json:"blocker"`
	Attacker string `json:"attacker"`
}

func blockSetParams(t *testing.T, entries ...blockerEntry) json.RawMessage {
	t.Helper()
	return params(t, map[string]any{"blocks": entries})
}

// pushMenacer puts a menace attacker on seat 0 and returns its ID.
func pushMenacer(t *testing.T, g *game.Game) string {
	t.Helper()
	c := game.NewCard("Menacer", g.Seats[0].ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 3, 3
	c.Keywords = []string{"menace"}
	g.Battlefield.PushTop(c)
	return c.InstanceID.String()
}

// menaceCombat parks a game in declare_blockers with a menace
// attacker from seat 0 and two untapped creatures on seat 1.
func menaceCombat(t *testing.T) (*game.Game, string, []string) {
	t.Helper()
	g := newGame(t)
	attacker := pushMenacer(t, g)
	blockers := []string{pushCreature(t, g, g.Seats[1]), pushCreature(t, g, g.Seats[1])}
	advanceTo(t, g, game.StepDeclareAttackers)
	act := mustAction(t, TypeDeclareAttacker, params(t, map[string]string{
		"attacker": attacker,
		"target":   g.Seats[1].ID.String(),
	}))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, attacker, blockers
}

func blockingTargetOf(g *game.Game, id string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == id {
			return c.BlockingTarget
		}
	}
	return uuid.Nil
}

func TestDispatchDeclareBlockersSet(t *testing.T) {
	g, attacker, blockers := menaceCombat(t)

	act := mustAction(t, TypeDeclareBlockers, blockSetParams(t,
		blockerEntry{Blocker: blockers[0], Attacker: attacker},
		blockerEntry{Blocker: blockers[1], Attacker: attacker},
	))
	act.Caller = g.Seats[1].ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, b := range blockers {
		if got := blockingTargetOf(g, b); got.String() != attacker {
			t.Errorf("%s is blocking %v, want %s", b, got, attacker)
		}
	}
}

// The one-entry path through the same verb is still refused when the
// count says so, and nothing is stored.
func TestDispatchDeclareBlockerAloneOnMenaceIsRefused(t *testing.T) {
	g, attacker, blockers := menaceCombat(t)

	act := mustAction(t, TypeDeclareBlocker, params(t, map[string]string{
		"blocker":  blockers[0],
		"attacker": attacker,
	}))
	act.Caller = g.Seats[1].ID
	if err := Dispatch(g, act); !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("Dispatch err = %v, want an illegal-block refusal", err)
	}
	if got := blockingTargetOf(g, blockers[0]); got != uuid.Nil {
		t.Errorf("a refused block was stored: blocking %v", got)
	}
}

func TestDispatchDeclareBlockersEmptySet(t *testing.T) {
	g, _, _ := menaceCombat(t)
	act := mustAction(t, TypeDeclareBlockers, blockSetParams(t))
	act.Caller = g.Seats[1].ID
	if err := Dispatch(g, act); !errors.Is(err, ErrEmptyBlockerSet) {
		t.Fatalf("Dispatch err = %v, want ErrEmptyBlockerSet", err)
	}
}

func TestDispatchDeclareBlockersTooMany(t *testing.T) {
	g, attacker, blockers := menaceCombat(t)
	entries := make([]blockerEntry, MaxBulkBlockers+1)
	for i := range entries {
		entries[i] = blockerEntry{Blocker: blockers[0], Attacker: attacker}
	}
	act := mustAction(t, TypeDeclareBlockers, blockSetParams(t, entries...))
	act.Caller = g.Seats[1].ID
	if err := Dispatch(g, act); !errors.Is(err, ErrTooManyBlockers) {
		t.Fatalf("Dispatch err = %v, want ErrTooManyBlockers", err)
	}
}

// Authorization rejects the whole batch, exactly as the single-card
// verb does: a creature the caller does not control is a permission
// question, not an eligibility one.
func TestDispatchDeclareBlockersRejectsSomeoneElsesCreature(t *testing.T) {
	g, attacker, blockers := menaceCombat(t)
	act := mustAction(t, TypeDeclareBlockers, blockSetParams(t,
		blockerEntry{Blocker: blockers[0], Attacker: attacker},
		blockerEntry{Blocker: blockers[1], Attacker: attacker},
	))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); err == nil {
		t.Fatal("blocking with another seat's creatures was allowed")
	}
	if got := blockingTargetOf(g, blockers[0]); got != uuid.Nil {
		t.Errorf("a rejected batch stored an entry anyway: blocking %v", got)
	}
}

func TestDispatchDeclareBlockersBadUUID(t *testing.T) {
	g, attacker, _ := menaceCombat(t)
	act := mustAction(t, TypeDeclareBlockers, blockSetParams(t,
		blockerEntry{Blocker: "not-a-uuid", Attacker: attacker},
	))
	act.Caller = g.Seats[1].ID
	if err := Dispatch(g, act); err == nil {
		t.Fatal("a malformed blocker ID was accepted")
	}
}

// #1279: finish_blocks completes the caller's own declaration and
// nobody else's.
func TestDispatchFinishBlocks(t *testing.T) {
	g, _, _ := menaceCombat(t)
	def := g.Seats[1].ID

	other, err := Decode(string(TypeFinishBlocks), def.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	other.Caller = g.Seats[0].ID
	if err := Dispatch(g, other); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("finishing another seat's declaration: %v, want ErrPlayerCallerMismatch", err)
	}
	if got := g.BlockDeclarationStatusOf(def); got != game.BlockDeclarationPending {
		t.Fatalf("status %q after a refused finish, want pending", got)
	}

	own, err := Decode(string(TypeFinishBlocks), def.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	own.Caller = def
	if err := Dispatch(g, own); err != nil {
		t.Fatalf("finish_blocks: %v", err)
	}
	if got := g.BlockDeclarationStatusOf(def); got != game.BlockDeclarationDeclared {
		t.Errorf("status %q, want declared", got)
	}
}
