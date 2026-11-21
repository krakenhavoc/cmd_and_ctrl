package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_chain_veil_test.go — The Chain Veil (#4191, Guffster's Party
// #4184).

// chainVeil puts an untapped Veil under p and four colourless mana in
// p's pool for its ability.
func chainVeil(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	veil := effectstest.PushCatalogPermanent(g, p.ID, "The Chain Veil", "Legendary Artifact", effectstest.GuffbChainVeilOracle, false)
	chainVeilMana(t, g, p)
	return veil
}

func chainVeilMana(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
}

// chainVeilActivate activates the Veil's {4}, {T} ability and lets it
// resolve.
func chainVeilActivate(t *testing.T, g *game.Game, p *game.Player, veil uuid.UUID) {
	t.Helper()
	effectstest.B16Activate(t, g, p.ID, veil, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

// Each of your planeswalkers may activate once more after the Veil
// resolves, and the gate, the enumerator and the wire agree; a third
// activation is still refused.
func TestChainVeilLetsEachOfYourWalkersActivateOnceMore(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	c := effectstest.GuffbChandra(g, me.ID)
	// A second walker with Chandra's rows under another name, so the
	// legend rule leaves both.
	other := effectstest.GuffbWalker(g, me.ID, "Second Walker", "Planeswalker — Chandra", effectstest.GuffbChandraOracle, 4)
	effectstest.ChandraPlusOne(t, g, me, c)
	effectstest.ChandraPlusOne(t, g, me, other)
	if effectstest.LoyaltyRowsOffered(g, me, c) != 0 || !effectstest.WalkerGreyed(t, g, me, c) {
		t.Fatal("setup: a spent walker is still offered")
	}

	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	if !effectstest.B16Tapped(t, g, veil) {
		t.Error("the Veil did not tap")
	}
	if effectstest.LoyaltyRowsOffered(g, me, c) == 0 {
		t.Error("the enumerator does not offer the Veil's extra activation")
	}
	if effectstest.WalkerGreyed(t, g, me, c) {
		t.Error("the wire still greys the walker after the Veil resolved")
	}
	effectstest.ChandraPlusOne(t, g, me, c)
	effectstest.ChandraPlusOne(t, g, me, other)
	if err := g.ActivateCatalogAbility(me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation was allowed under one Veil")
	}
	if effectstest.LoyaltyRowsOffered(g, me, c) != 0 || !effectstest.WalkerGreyed(t, g, me, c) {
		t.Error("the walker is still offered after its Veil activation")
	}
	if got := effectstest.LoyaltyOf(g, c); got != 6 {
		t.Errorf("loyalty %d, want 6 after two +1s", got)
	}
}

// The ruling: the Veil modifies the rules, so a planeswalker that comes
// under your control after it resolves may activate twice too.
func TestChainVeilCoversAWalkerThatEntersAfterIt(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.ChandraPlusOne(t, g, me, c)
	effectstest.ChandraPlusOne(t, g, me, c)
	if err := g.ActivateCatalogAbility(me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation was allowed")
	}
}

// Untapped and activated again, the Veil gives a third activation; with
// Oath of Teferi, one Veil makes three as well.
func TestChainVeilStacksWithItselfAndWithOathOfTeferi(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	c := effectstest.GuffbChandra(g, me.ID)
	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	g.WithWriteLock(func() {
		if err := g.UntapTargetForEffect(veil); err != nil {
			t.Fatalf("untap the Veil: %v", err)
		}
	})
	chainVeilMana(t, g, me)
	chainVeilActivate(t, g, me, veil)
	for i := 1; i <= 3; i++ {
		effectstest.ChandraPlusOne(t, g, me, c)
	}
	if err := g.ActivateCatalogAbility(me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a fourth activation was allowed under two Veils")
	}

	g2, me2, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g2, me2.ID, "Oath of Teferi", "Legendary Enchantment", effectstest.GuffbOathOracle, false)
	c2 := effectstest.GuffbChandra(g2, me2.ID)
	effectstest.ChandraPlusOne(t, g2, me2, c2)
	veil2 := chainVeil(t, g2, me2)
	chainVeilActivate(t, g2, me2, veil2)
	effectstest.ChandraPlusOne(t, g2, me2, c2)
	effectstest.ChandraPlusOne(t, g2, me2, c2)
	if err := g2.ActivateCatalogAbility(me2.ID, c2, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a fourth activation was allowed under Oath and one Veil")
	}
}

// "As though none of its loyalty abilities have been activated": a
// walker that made its two under Oath of Teferi and then lost the Oath
// still gets the Veil's one.
func TestChainVeilGrantOutlastsAnOathThatLeaves(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	oath := effectstest.PushCatalogPermanent(g, me.ID, "Oath of Teferi", "Legendary Enchantment", effectstest.GuffbOathOracle, false)
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.ChandraPlusOne(t, g, me, c)
	effectstest.ChandraPlusOne(t, g, me, c)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(oath); err != nil {
			t.Fatalf("destroy the Oath: %v", err)
		}
	})
	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	effectstest.ChandraPlusOne(t, g, me, c)
	if err := g.ActivateCatalogAbility(me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a fourth activation was allowed")
	}
}

// At the beginning of your end step you lose 2 life unless you
// activated a loyalty ability this turn.
func TestChainVeilEndStepLosesTwoWithoutALoyaltyActivation(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "The Chain Veil", "Legendary Artifact", effectstest.GuffbChainVeilOracle, false)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before-2 {
		t.Errorf("life %d, want %d", got, before-2)
	}
}

func TestChainVeilEndStepCostsNothingAfterALoyaltyActivation(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "The Chain Veil", "Legendary Artifact", effectstest.GuffbChainVeilOracle, false)
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.ChandraPlusOne(t, g, me, c)
	// It doesn't matter that the walker has gone by the end step.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(c); err != nil {
			t.Fatalf("destroy Chandra: %v", err)
		}
	})
	before := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before {
		t.Errorf("life %d, want %d", got, before)
	}
}

// Only your own end step: an opponent's Veil does not cost you life.
func TestChainVeilEndStepIsOnlyYours(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "The Chain Veil", "Legendary Artifact", effectstest.GuffbChainVeilOracle, false)
	mine, theirs := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, me.ID) != mine || effectstest.LifeOf(g, opp.ID) != theirs {
		t.Errorf("life %d / %d, want %d / %d on the other player's end step", effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID), mine, theirs)
	}
}
