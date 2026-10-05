package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// combat_restrictions_2171_test.go — Bloodthirster (#2171): a creature
// that can't attack a player it has already attacked this turn.

const bloodthirsterOracle = "e971249a-64a3-4a9b-9a0c-e9d858ca8a55"

// bloodthirsterTargets is who the enumerator offers `attacker` at, as
// seat ids.
func bloodthirsterTargets(t *testing.T, g *game.Game, seat, attacker uuid.UUID) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindAttack || m.Source != attacker {
			continue
		}
		var p struct {
			Target string `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("attack params %s: %v", m.Params, err)
		}
		out[p.Target] = true
	}
	return out
}

func pushBloodthirster(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Bloodthirster", "Creature — Demon", bloodthirsterOracle, false)
}

// An extra combat: the first attack on opp1 is on the turn's record, and
// in the second combat it may attack only the players it has not.
func TestBloodthirsterExtraCombatOffersOnlyUnattackedPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, p1, p2, p3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	brute := pushBloodthirster(g, me.ID)
	advanceToDeclareAttackersOf(t, g, 0)

	first := bloodthirsterTargets(t, g, me.ID, brute)
	if len(first) != 3 {
		t.Fatalf("first combat offers %v, want all three opponents", first)
	}
	if err := g.DeclareAttacker(brute, p1.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	lockInAttacks(t, g)

	g.WithWriteLock(func() {
		g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})
	for i := 0; i < 60 && !(g.Turn.Step == game.StepDeclareAttackers && g.Turn.PhaseOrdinal == 2); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.PhaseOrdinal != 2 || g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("never reached the extra combat (step %s, phase %d)", g.Turn.Step, g.Turn.PhaseOrdinal)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, brute).Tapped = false })

	second := bloodthirsterTargets(t, g, me.ID, brute)
	if second[p1.ID.String()] || !second[p2.ID.String()] || !second[p3.ID.String()] || len(second) != 2 {
		t.Fatalf("extra combat offers %v, want only the two unattacked opponents", second)
	}
	err := g.Clone().DeclareAttacker(brute, p1.ID)
	var refused *game.AttackTargetRestrictionError
	if !errors.As(err, &refused) || !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("re-attacking the same player: err = %v, want an attack-target restriction refusal", err)
	}
	if err := g.DeclareAttacker(brute, p2.ID); err != nil {
		t.Fatalf("attacking a different player: %v", err)
	}
}

// Once it has attacked everyone it has nothing left to attack, and the
// engine and the enumerator agree on that.
func TestBloodthirsterCannotAttackOnceItHasAttackedEveryone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	brute := pushBloodthirster(g, me.ID)
	advanceToDeclareAttackersOf(t, g, 0)
	g.WithWriteLock(func() {
		for _, p := range g.Seats[1:] {
			g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: brute, Actor: me.ID, Target: p.ID})
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if got := bloodthirsterTargets(t, g, me.ID, brute); len(got) != 0 {
		t.Fatalf("offered %v after attacking everyone", got)
	}
	for _, p := range g.Seats[1:] {
		if err := g.Clone().DeclareAttacker(brute, p.ID); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("attack at %s: err = %v, want ErrIllegalAttackTarget", p.ID, err)
		}
	}
}

// CR 400.7: a flickered Bloodthirster is a new object and has attacked
// nobody.
func TestBloodthirsterNewObjectMayAttackAgain(t *testing.T) {
	g := newCatalogGame(t)
	me, p1 := g.Seats[0], g.Seats[1]
	brute := pushBloodthirster(g, me.ID)
	advanceToDeclareAttackersOf(t, g, 0)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: brute, Actor: me.ID, Target: p1.ID})
	})
	if bloodthirsterTargets(t, g, me.ID, brute)[p1.ID.String()] {
		t.Fatal("setup: the attacked player is still offered")
	}
	flickerInResponse(t, g, brute)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, brute).SummonedThisTurn = false })
	if !bloodthirsterTargets(t, g, me.ID, brute)[p1.ID.String()] {
		t.Fatal("a new object is offered the player the old one attacked")
	}
	if err := g.Clone().DeclareAttacker(brute, p1.ID); err != nil {
		t.Fatalf("a new object attacking the same player: %v", err)
	}
}
