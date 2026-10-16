package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vines_of_vastwood_test.go — #3860. "Target creature can't be the
// target of spells or abilities YOUR OPPONENTS control this turn" is not
// hexproof: it protects against the caster's opponents, whoever controls
// the creature. The first cut narrowed the card to "a creature you
// control" so hexproof was exact, which refused the Vines on anyone
// else's creature.

// castVinesAt casts Vines of Vastwood for `caster` at `target` and lets
// it resolve.
func castVinesAt(t *testing.T, g *game.Game, caster *game.Player, target uuid.UUID) {
	t.Helper()
	id := pushCatalogHandCard(caster, "Vines of Vastwood", "Instant", vinesOfVastwoodOracle)
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{Targets: cardTarget(target)}); err != nil {
		t.Fatalf("cast Vines of Vastwood: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// boltFromHand casts a Lightning Bolt from `p`'s hand at `target`,
// returning the cast's error.
func boltFromHand(g *game.Game, p *game.Player, target game.TargetRef) error {
	id := pushCatalogHandCard(p, "Lightning Bolt", "Instant", lightningBoltOracle)
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: []game.TargetRef{target}})
}

func TestVinesOfVastwoodStopsTheCastersOpponentsTargetingTheirCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castVinesAt(t, g, me, mine)

	if err := boltFromHand(g, opp, game.TargetRef{Kind: game.TargetCard, ID: mine}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an opponent's Bolt at the protected creature: got %v, want ErrIllegalTarget", err)
	}
	// The caster is not their own opponent: they can still aim at it.
	if err := boltFromHand(g, me, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Errorf("the caster can still target the creature: %v", err)
	}
	// And hexproof is not what it is: the keyword is not granted.
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("the clause is a targeting bar, not the hexproof keyword")
	}
}

// The reason the card is not hexproof: cast on a creature an OPPONENT
// controls, it locks that creature's own controller out of targeting it
// (and every other opponent of the caster), while the caster still can.
func TestVinesOfVastwoodOnAnOpponentsCreatureBarsThatOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castVinesAt(t, g, me, theirs) // legal now: any creature

	if err := boltFromHand(g, opp, game.TargetRef{Kind: game.TargetCard, ID: theirs}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("the creature's own controller is the caster's opponent: got %v, want ErrIllegalTarget", err)
	}
	if err := boltFromHand(g, other, game.TargetRef{Kind: game.TargetCard, ID: theirs}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a third player is the caster's opponent too: got %v, want ErrIllegalTarget", err)
	}
	if err := boltFromHand(g, me, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Errorf("the caster is not barred: %v", err)
	}
}

// A spell already aimed at the creature when the Vines resolve fizzles
// at resolution (CR 608.2b): the bar is read by the same gate that
// re-checks a target.
func TestVinesOfVastwoodInResponseFizzlesTheSpellAlreadyAimedAtIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	if err := boltFromHand(g, opp, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Fatalf("setup: the Bolt is legal before the Vines: %v", err)
	}
	vines := pushCatalogHandCard(me, "Vines of Vastwood", "Instant", vinesOfVastwoodOracle)
	if err := g.CastSpell(me.ID, vines, game.CastSpellParams{Targets: cardTarget(mine)}); err != nil {
		t.Fatalf("cast Vines in response: %v", err)
	}
	passPriorityAroundTable(t, g)

	if damageMarkedOn(g, mine) != 0 || !g.Battlefield.Contains(mine) {
		t.Errorf("the Bolt should have fizzled: damage %d, on battlefield %v", damageMarkedOn(g, mine), g.Battlefield.Contains(mine))
	}
}

// The bar ends at cleanup with the turn.
func TestVinesOfVastwoodEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castVinesAt(t, g, me, mine)
	advanceToMainOf(t, g, 1)
	if err := boltFromHand(g, opp, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Errorf("next turn the creature can be targeted again: %v", err)
	}
}

// "As though it didn't have hexproof" waives hexproof and nothing else
// (hexproof_bypass.go): a bar is not hexproof, so a waiver cannot lift
// it. Pinned at the gate, with a waiver record on the opponent.
func TestVinesOfVastwoodIsNotLiftedByAHexproofWaiver(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castVinesAt(t, g, me, mine)
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, game.ScopeOpponentsAndTheirCreatures, opp.ID,
			[]game.Mod{game.WaiveHexproofMod()}, g.UntilEndOfTurnDuration(), "test waiver")
	})
	if err := boltFromHand(g, opp, game.TargetRef{Kind: game.TargetCard, ID: mine}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a hexproof waiver does not lift the bar: got %v", err)
	}
}

func TestVinesOfVastwoodIsComplete(t *testing.T) {
	spec, ok := Lookup(vinesOfVastwoodOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Errorf("Vines of Vastwood: %+v, want a Full catalog entry (#3860)", spec.Completeness)
	}
}
