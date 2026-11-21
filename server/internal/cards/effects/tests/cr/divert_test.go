package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDivertPaidChangesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	payer, victim := g.Seats[0], g.Seats[2].ID
	for i := 0; i < 2; i++ {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Island",
			TypeLine: "Basic Land — Island", Owner: payer.ID, Controller: payer.ID,
		})
	}
	bolt, _ := effectstest.BoltThenDivert(t, g)
	effectstest.PassUntilTaxed(t, g, payer.ID)
	effectstest.AnswerPayUnless(t, g, payer.ID, true)
	if effectstest.LatestRetarget(g, g.Seats[1].ID) != nil {
		t.Fatal("a paid Divert still offered a retarget")
	}
	if got := g.StackMeta[bolt].Targets; len(got) != 1 || got[0].ID != victim {
		t.Fatalf("a paid Divert moved the target: %+v", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victim); got != 37 {
		t.Errorf("victim life = %d, want 37 (the Bolt resolved where it was aimed)", got)
	}
}

func TestDivertDeclinedRetargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	payer, caster := g.Seats[0], g.Seats[1]
	victim, other := g.Seats[2].ID, g.Seats[3].ID
	bolt, _ := effectstest.BoltThenDivert(t, g)
	effectstest.PassUntilTaxed(t, g, payer.ID)
	effectstest.AnswerPayUnless(t, g, payer.ID, false)

	prompt := effectstest.LatestRetarget(g, caster.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victim) != 40 || effectstest.LifeOf(g, other) != 37 {
		t.Errorf("damage landed wrong: old=%d new=%d", effectstest.LifeOf(g, victim), effectstest.LifeOf(g, other))
	}
}

// The guarded spell must not resolve before its controller answers.
func TestDivertHoldsTheGuardedSpellUntilAnswered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	payer, victim := g.Seats[0], g.Seats[2].ID
	bolt, _ := effectstest.BoltThenDivert(t, g)
	effectstest.PassUntilTaxed(t, g, payer.ID)

	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a pass while Divert's tax was open = %v, want ErrChoicePending", err)
	}
	if !g.Stack.Contains(bolt) {
		t.Fatal("the Bolt resolved while Divert's question was unanswered")
	}
	if got := effectstest.LifeOf(g, victim); got != 40 {
		t.Errorf("victim life = %d before the answer, want 40", got)
	}
}

// A spell that is no longer on the stack when Divert resolves raises
// no prompt and nothing happens.
func TestDivertOnAGoneSpellDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	payer := g.Seats[0]
	bolt, _ := effectstest.BoltThenDivert(t, g)
	g.WithWriteLock(func() {
		for i, c := range g.Stack.Cards {
			if c.InstanceID == bolt {
				g.Stack.Cards = append(g.Stack.Cards[:i], g.Stack.Cards[i+1:]...)
				break
			}
		}
		delete(g.StackMeta, bolt)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, payer.ID) {
		t.Error("Divert asked about a spell that is gone")
	}
}

// A spell with two targets is not a legal Divert target.
func TestDivertRefusesAMultiTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimB := g.Seats[0], g.Seats[2].ID
	double := effectstest.CastCatalogSpell(t, g, "Bite Down", "Instant", effectstest.B26BiteDownOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: effectstest.PushPermanent(g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Test", Power: 5, Toughness: 5}), Slot: 0},
		{Kind: game.TargetCard, ID: effectstest.PushPermanent(g, victimB, game.Card{Name: "Theirs", TypeLine: "Creature — Test", Power: 2, Toughness: 2}), Slot: 1},
	})
	caster := g.Seats[1]
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Divert", TypeLine: "Instant",
		OracleID: effectstest.DivertOracle, Owner: caster.ID, Controller: caster.ID})
	err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: double}},
	})
	if err == nil {
		t.Error("Divert accepted a two-target spell")
	}
}
