package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divert_test.go — #2365: the guarded pay-unless with a custom decline.

const divertOracle = "561aed01-322b-4699-b7b0-f4416075aa9b"

// boltThenDivert has the active seat (the future payer) Bolt seat 2 and
// seat 1 answer with Divert. Seat 0 is asked once Divert resolves.
func boltThenDivert(t *testing.T, g *game.Game) (bolt, divert uuid.UUID) {
	t.Helper()
	victim := g.Seats[2].ID
	bolt = castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	caster := g.Seats[1]
	divert = uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: divert, Name: "Divert", TypeLine: "Instant", OracleID: divertOracle,
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, divert, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("CastSpell Divert: %v", err)
	}
	return bolt, divert
}

func TestDivertPaidChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	payer, victim := g.Seats[0], g.Seats[2].ID
	for i := 0; i < 2; i++ {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Island",
			TypeLine: "Basic Land — Island", Owner: payer.ID, Controller: payer.ID,
		})
	}
	bolt, _ := boltThenDivert(t, g)
	passUntilTaxed(t, g, payer.ID)
	answerPayUnless(t, g, payer.ID, true)
	if latestRetarget(g, g.Seats[1].ID) != nil {
		t.Fatal("a paid Divert still offered a retarget")
	}
	if got := g.StackMeta[bolt].Targets; len(got) != 1 || got[0].ID != victim {
		t.Fatalf("a paid Divert moved the target: %+v", got)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim); got != 37 {
		t.Errorf("victim life = %d, want 37 (the Bolt resolved where it was aimed)", got)
	}
}

func TestDivertDeclinedRetargets(t *testing.T) {
	g := newCatalogGame(t)
	payer, caster := g.Seats[0], g.Seats[1]
	victim, other := g.Seats[2].ID, g.Seats[3].ID
	bolt, _ := boltThenDivert(t, g)
	passUntilTaxed(t, g, payer.ID)
	answerPayUnless(t, g, payer.ID, false)

	prompt := latestRetarget(g, caster.ID)
	if prompt == nil {
		t.Fatal("a declined Divert opened no retarget prompt for its controller")
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("Divert's change is mandatory; prompt min = %d", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, caster.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: other}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[bolt].Targets; got[0].ID != other {
		t.Fatalf("the Bolt's target did not move: %+v", got)
	}
	passPriorityAroundTable(t, g)
	if lifeOf(g, victim) != 40 || lifeOf(g, other) != 37 {
		t.Errorf("damage landed wrong: old=%d new=%d", lifeOf(g, victim), lifeOf(g, other))
	}
}

// The guarded spell must not resolve before its controller answers.
func TestDivertHoldsTheGuardedSpellUntilAnswered(t *testing.T) {
	g := newCatalogGame(t)
	payer, victim := g.Seats[0], g.Seats[2].ID
	bolt, _ := boltThenDivert(t, g)
	passUntilTaxed(t, g, payer.ID)

	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while Divert's tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(bolt) {
		t.Fatal("the Bolt resolved while Divert's question was unanswered")
	}
	if got := lifeOf(g, victim); got != 40 {
		t.Errorf("victim life = %d before the answer, want 40", got)
	}
}

// A spell that is no longer on the stack when Divert resolves raises
// no prompt and nothing happens.
func TestDivertOnAGoneSpellDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	payer := g.Seats[0]
	bolt, _ := boltThenDivert(t, g)
	g.WithWriteLock(func() {
		for i, c := range g.Stack.Cards {
			if c.InstanceID == bolt {
				g.Stack.Cards = append(g.Stack.Cards[:i], g.Stack.Cards[i+1:]...)
				break
			}
		}
		delete(g.StackMeta, bolt)
	})
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, payer.ID) {
		t.Error("Divert asked about a spell that is gone")
	}
}

// A spell with two targets is not a legal Divert target.
func TestDivertRefusesAMultiTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victimB := g.Seats[0], g.Seats[2].ID
	double := castCatalogSpell(t, g, "Bite Down", "Instant", b26BiteDownOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: pushPermanent(g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Test", Power: 5, Toughness: 5}), Slot: 0},
		{Kind: game.TargetCard, ID: pushPermanent(g, victimB, game.Card{Name: "Theirs", TypeLine: "Creature — Test", Power: 2, Toughness: 2}), Slot: 1},
	})
	caster := g.Seats[1]
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Divert", TypeLine: "Instant",
		OracleID: divertOracle, Owner: caster.ID, Controller: caster.ID})
	err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: double}},
	})
	if err == nil {
		t.Error("Divert accepted a two-target spell")
	}
}
