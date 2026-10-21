package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// the_chain_veil_test.go — The Chain Veil (#4191, Guffster's Party
// #4184).

// chainVeil puts an untapped Veil under p and four colourless mana in
// p's pool for its ability.
func chainVeil(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	veil := pushCatalogPermanent(g, p.ID, "The Chain Veil", "Legendary Artifact", guffbChainVeilOracle, false)
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
	b16Activate(t, g, p.ID, veil, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
}

// chandraPlusOne activates Chandra's "+1: Add {R}{R}." and lets it
// resolve.
func chandraPlusOne(t *testing.T, g *game.Game, p *game.Player, c uuid.UUID) {
	t.Helper()
	b16Activate(t, g, p.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
}

// loyaltyRowsOffered counts the enumerator's loyalty-row moves for
// `walker`.
func loyaltyRowsOffered(g *game.Game, p *game.Player, walker uuid.UUID) int {
	n := 0
	for _, m := range legal.EnumerateFor(g, p.ID) {
		if m.Type == "activate_ability" && m.Source == walker {
			n++
		}
	}
	return n
}

// walkerGreyed is the wire's loyalty_activated flag for `walker`, as
// p's client sees it.
func walkerGreyed(t *testing.T, g *game.Game, p *game.Player, walker uuid.UUID) bool {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, s := range g.Seats {
				c.KnownBy[s.ID] = true
			}
		}
	})
	v := protocol.ViewOfGameFor(g, p.ID.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID == walker.String() {
			return cv.LoyaltyActivated
		}
	}
	t.Fatalf("%s is not in the view", walker)
	return false
}

// Each of your planeswalkers may activate once more after the Veil
// resolves, and the gate, the enumerator and the wire agree; a third
// activation is still refused.
func TestChainVeilLetsEachOfYourWalkersActivateOnceMore(t *testing.T) {
	g, me, _ := spendTable(t)
	c := guffbChandra(g, me.ID)
	// A second walker with Chandra's rows under another name, so the
	// legend rule leaves both.
	other := guffbWalker(g, me.ID, "Second Walker", "Planeswalker — Chandra", guffbChandraOracle, 4)
	chandraPlusOne(t, g, me, c)
	chandraPlusOne(t, g, me, other)
	if loyaltyRowsOffered(g, me, c) != 0 || !walkerGreyed(t, g, me, c) {
		t.Fatal("setup: a spent walker is still offered")
	}

	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	if !b16Tapped(t, g, veil) {
		t.Error("the Veil did not tap")
	}
	if loyaltyRowsOffered(g, me, c) == 0 {
		t.Error("the enumerator does not offer the Veil's extra activation")
	}
	if walkerGreyed(t, g, me, c) {
		t.Error("the wire still greys the walker after the Veil resolved")
	}
	chandraPlusOne(t, g, me, c)
	chandraPlusOne(t, g, me, other)
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation was allowed under one Veil")
	}
	if loyaltyRowsOffered(g, me, c) != 0 || !walkerGreyed(t, g, me, c) {
		t.Error("the walker is still offered after its Veil activation")
	}
	if got := loyaltyOf(g, c); got != 6 {
		t.Errorf("loyalty %d, want 6 after two +1s", got)
	}
}

// The ruling: the Veil modifies the rules, so a planeswalker that comes
// under your control after it resolves may activate twice too.
func TestChainVeilCoversAWalkerThatEntersAfterIt(t *testing.T) {
	g, me, _ := spendTable(t)
	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	c := guffbChandra(g, me.ID)
	chandraPlusOne(t, g, me, c)
	chandraPlusOne(t, g, me, c)
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation was allowed")
	}
}

// Untapped and activated again, the Veil gives a third activation; with
// Oath of Teferi, one Veil makes three as well.
func TestChainVeilStacksWithItselfAndWithOathOfTeferi(t *testing.T) {
	g, me, _ := spendTable(t)
	c := guffbChandra(g, me.ID)
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
		chandraPlusOne(t, g, me, c)
	}
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a fourth activation was allowed under two Veils")
	}

	g2, me2, _ := spendTable(t)
	pushCatalogPermanent(g2, me2.ID, "Oath of Teferi", "Legendary Enchantment", guffbOathOracle, false)
	c2 := guffbChandra(g2, me2.ID)
	chandraPlusOne(t, g2, me2, c2)
	veil2 := chainVeil(t, g2, me2)
	chainVeilActivate(t, g2, me2, veil2)
	chandraPlusOne(t, g2, me2, c2)
	chandraPlusOne(t, g2, me2, c2)
	if err := g2.ActivateCatalogAbility(me2.ID, c2, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a fourth activation was allowed under Oath and one Veil")
	}
}

// "As though none of its loyalty abilities have been activated": a
// walker that made its two under Oath of Teferi and then lost the Oath
// still gets the Veil's one.
func TestChainVeilGrantOutlastsAnOathThatLeaves(t *testing.T) {
	g, me, _ := spendTable(t)
	oath := pushCatalogPermanent(g, me.ID, "Oath of Teferi", "Legendary Enchantment", guffbOathOracle, false)
	c := guffbChandra(g, me.ID)
	chandraPlusOne(t, g, me, c)
	chandraPlusOne(t, g, me, c)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(oath); err != nil {
			t.Fatalf("destroy the Oath: %v", err)
		}
	})
	veil := chainVeil(t, g, me)
	chainVeilActivate(t, g, me, veil)
	chandraPlusOne(t, g, me, c)
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a fourth activation was allowed")
	}
}

// At the beginning of your end step you lose 2 life unless you
// activated a loyalty ability this turn.
func TestChainVeilEndStepLosesTwoWithoutALoyaltyActivation(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "The Chain Veil", "Legendary Artifact", guffbChainVeilOracle, false)
	before := lifeOf(g, me.ID)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != before-2 {
		t.Errorf("life %d, want %d", got, before-2)
	}
}

func TestChainVeilEndStepCostsNothingAfterALoyaltyActivation(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "The Chain Veil", "Legendary Artifact", guffbChainVeilOracle, false)
	c := guffbChandra(g, me.ID)
	chandraPlusOne(t, g, me, c)
	// It doesn't matter that the walker has gone by the end step.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(c); err != nil {
			t.Fatalf("destroy Chandra: %v", err)
		}
	})
	before := lifeOf(g, me.ID)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != before {
		t.Errorf("life %d, want %d", got, before)
	}
}

// Only your own end step: an opponent's Veil does not cost you life.
func TestChainVeilEndStepIsOnlyYours(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, opp.ID, "The Chain Veil", "Legendary Artifact", guffbChainVeilOracle, false)
	mine, theirs := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if lifeOf(g, me.ID) != mine || lifeOf(g, opp.ID) != theirs {
		t.Errorf("life %d / %d, want %d / %d on the other player's end step", lifeOf(g, me.ID), lifeOf(g, opp.ID), mine, theirs)
	}
}
