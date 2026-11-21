package testscm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castCounterExile has seat 1 answer `victim` (seat 0's spell on the
// stack) with Syncopate at `x` or, for x < 0, No More Lies, and passes
// until seat 0 owes the tax.
func castCounterExile(t *testing.T, g *game.Game, victim uuid.UUID, x int) *game.PendingChoice {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	if x < 0 {
		effectstest.CastFromSeat(t, g, opp, "No More Lies", "Instant", effectstest.NoMoreLiesOracle, targets)
	} else {
		id := effectstest.HandCardFull(opp, "Syncopate", "Instant", "{X}{U}", effectstest.SyncopateOracle, []string{"U"})
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{XValue: x, Targets: targets}); err != nil {
			t.Fatalf("casting Syncopate at X=%d: %v", x, err)
		}
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	c := effectstest.PayUnlessFor(g, me.ID)
	if c.PickThen != effectstest.CounterExileKey || c.GuardsStackItem != victim {
		t.Fatalf("the tax is %+v, want the keyed exiling counter guarding the spell", c)
	}
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the tax was unanswered")
	}
	return c
}

// addColorless puts n colourless mana in p's pool.
func addColorless(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

// assertExiled checks that a declined tax exiled the spell rather than
// putting it into its owner's graveyard.
func assertExiled(t *testing.T, g *game.Game, owner *game.Player, spell uuid.UUID) {
	t.Helper()
	if g.Stack.Contains(spell) {
		t.Fatal("the spell is still on the stack after the tax was declined")
	}
	if !g.Exile.Contains(spell) || owner.Graveyard.Contains(spell) {
		t.Errorf("exiled=%v in graveyard=%v, want it exiled instead of put into the graveyard",
			g.Exile.Contains(spell), owner.Graveyard.Contains(spell))
	}
}

// assertResolved checks that a paid tax let the spell resolve: a
// sorcery goes to its owner's graveyard as it resolves, and nothing is
// exiled.
func assertResolved(t *testing.T, g *game.Game, owner *game.Player, spell uuid.UUID) {
	t.Helper()
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell left the stack although the tax was paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || g.Exile.Contains(spell) || !owner.Graveyard.Contains(spell) {
		t.Errorf("after resolving: on stack=%v exiled=%v in graveyard=%v, want it resolved into the graveyard",
			g.Stack.Contains(spell), g.Exile.Contains(spell), owner.Graveyard.Contains(spell))
	}
}

// Unpaid, the spell is exiled, not put into its owner's graveyard.
func TestNoMoreLiesDeclinedExilesTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	c := castCounterExile(t, g, victim, -1)
	if c.PayCost != "{3}" {
		t.Errorf("No More Lies asks for %q, want {3}", c.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	assertExiled(t, g, me, victim)
}

// Paid, the spell resolves as normal.
func TestNoMoreLiesPaidLetsTheSpellResolve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	castCounterExile(t, g, victim, -1)
	addColorless(me, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	assertResolved(t, g, me, victim)
}

// A restore point taken while the payer is deciding keeps the
// destination: declined on the restored table, the spell is still
// exiled, and paid it still resolves.
func TestNoMoreLiesTaxIsARestorePointThatStillExiles(t *testing.T) {
	for _, pay := range []bool{false, true} {
		t.Run(fmt.Sprintf("pay=%v", pay), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMainForCost(t, g)
			victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
			castCounterExile(t, g, victim, -1)

			g = effectstest.RestoreRoundTrip(t, g, true)
			me := g.Seats[0]
			c := effectstest.PayUnlessFor(g, me.ID)
			if c == nil || c.PickThen != effectstest.CounterExileKey || c.GuardsStackItem != victim || c.PayCost != "{3}" {
				t.Fatalf("restored: %+v, want the keyed exiling counter guarding the spell", c)
			}
			if err := effectstest.TheThirdSeatPasses(t, g); !errors.Is(err, game.ErrChoicePending) {
				t.Fatalf("a pass while the restored tax was open = %v, want ErrChoicePending", err)
			}
			if pay {
				addColorless(me, 3)
			}
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
			if pay {
				assertResolved(t, g, me, victim)
			} else {
				assertExiled(t, g, me, victim)
			}
		})
	}
}

// A spell that can't be countered is neither countered nor exiled by a
// declined tax: it stays on the stack and resolves.
func TestNoMoreLiesDoesNotExileASpellThatCantBeCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	verdict := effectstest.CastCatalogSpell(t, g, "Supreme Verdict", "Sorcery", effectstest.SupremeVerdictOracle, nil)
	castCounterExile(t, g, verdict, -1)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Exile.Contains(verdict) || me.Graveyard.Contains(verdict) {
		t.Fatal("a declined tax moved a spell that can't be countered")
	}
	if n := effectstest.CountEventsOf(g, game.EventCounterSpell, verdict); n != 0 {
		t.Errorf("%d EventCounterSpell for a spell that can't be countered", n)
	}
	assertResolved(t, g, me, verdict)
}

// Syncopate asks for exactly the X it was cast with, X = 0 included: at
// X = 0 the tax is {0}, which the controller pays with an empty pool. A
// "Pay" the controller cannot fund is a decline, and exiles.
func TestSyncopateReadsItsX(t *testing.T) {
	for _, tc := range []struct {
		x, mana int
		pay     bool
		exiled  bool
	}{
		{x: 0, pay: true},
		{x: 0, pay: false, exiled: true},
		{x: 2, mana: 2, pay: true},
		{x: 2, pay: false, exiled: true},
		{x: 2, mana: 1, pay: true, exiled: true},
		{x: 4, mana: 4, pay: true},
	} {
		t.Run(fmt.Sprintf("x=%d mana=%d pay=%v", tc.x, tc.mana, tc.pay), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.ToMainForCost(t, g)
			victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
			c := castCounterExile(t, g, victim, tc.x)
			if want := fmt.Sprintf("{%d}", tc.x); c.PayCost != want {
				t.Fatalf("Syncopate at X=%d asks for %q, want %q", tc.x, c.PayCost, want)
			}
			addColorless(me, tc.mana)
			effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			if tc.exiled {
				assertExiled(t, g, me, victim)
			} else {
				assertResolved(t, g, me, victim)
			}
		})
	}
}
