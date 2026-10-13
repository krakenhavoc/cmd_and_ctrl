package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_b_encore_test.go — encore (CR 702.141, #3172) on Impulsive Pilferer.

func pilfererTokens(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Impulsive Pilferer" && c.Controller == owner {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// {3}{R}, exile it from your graveyard: a hasty token copy for EACH
// opponent, each of which must attack its own opponent, and all of which
// are sacrificed at the beginning of the next end step.
func TestImpulsivePilfererEncoreMakesOneBoundTokenPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Impulsive Pilferer", "Creature — Goblin Pirate",
		impulsivePilfererOracle, 1, 1, "{C}{C}{C}{R}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("encore's cost exiles the card from the graveyard")
	}
	passPriorityAroundTable(t, g)
	toks := pilfererTokens(g, me.ID)
	if len(toks) != 3 {
		t.Fatalf("tokens: %d, want one per opponent (3)", len(toks))
	}
	for _, tok := range toks {
		if !hasEffectiveKeyword(t, g, tok, "haste") {
			t.Errorf("token %s has no haste", tok)
		}
	}

	// Each token is bound to a different opponent.
	bound := map[uuid.UUID]uuid.UUID{}
	g.ReadSnapshot(func() {
		for _, tok := range toks {
			c, _ := g.LookupCardForEffect(tok)
			for _, r := range c.Effective().AttackRequirements {
				if r.MustAttack != uuid.Nil {
					bound[tok] = r.MustAttack
				}
			}
		}
	})
	seen := map[uuid.UUID]bool{}
	for _, tok := range toks {
		o, ok := bound[tok]
		if !ok || o == me.ID {
			t.Fatalf("token %s is bound to %v, want one of the opponents", tok, o)
		}
		seen[o] = true
	}
	if len(seen) != 3 {
		t.Errorf("tokens are bound to %d distinct opponents, want 3", len(seen))
	}

	// Attack the wrong player with one: the pass is refused and names it.
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	wrong := toks[0]
	var other uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != bound[wrong] {
			other = s.ID
			break
		}
	}
	if err := g.DeclareAttacker(wrong, other); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("declaring a token at the wrong player = %v, want ErrAttackRequirement", err)
	}
	for _, tok := range toks {
		if err := g.DeclareAttacker(tok, bound[tok]); err != nil {
			t.Fatalf("declare each token at its own opponent: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with every token at its own opponent: %v", err)
	}

	// Sacrificed at the next end step.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if left := pilfererTokens(g, me.ID); len(left) != 0 {
		t.Errorf("%d tokens survived the end step", len(left))
	}
}

// A sorcery-speed ability: refused at instant speed.
func TestImpulsivePilfererEncoreIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepBeginCombat)
	id := pushCatalogGraveyardCard(me, "Impulsive Pilferer", "Creature — Goblin Pirate", impulsivePilfererOracle, 1, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("encore is activate-only-as-a-sorcery")
	}
	if me.Graveyard.Contains(id) == false {
		t.Error("a refused activation must not have exiled the card")
	}
}
