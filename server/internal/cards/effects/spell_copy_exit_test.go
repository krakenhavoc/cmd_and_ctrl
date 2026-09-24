package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_copy_exit_test.go — #1340, CR 707.10a / CR 704.5e on real
// cards. A copy of a spell that leaves the stack WITHOUT resolving —
// countered, bounced, exiled — ceases to exist. Before #1340 the
// resolution frame knew that and the counter / bounce / exile exit did
// not, so a countered Twincast copy was a real Lightning Bolt in a
// graveyard: countable by Tarmogoyf, returnable by Regrowth.
//
// The engine shape is pinned in game/spell_copy_exit_test.go. This file
// is one test per way a copy is made or removed that a player will
// actually do.

// copyOnStack reports the one spell COPY on the stack, or uuid.Nil.
func copyOnStack(g *game.Game) uuid.UUID {
	for id, item := range g.StackMeta {
		if item != nil && item.IsCopy && g.Stack.Contains(id) {
			return id
		}
	}
	return uuid.Nil
}

// anywhereButTheStack reports whether `id` is in any zone other than
// the stack — the thing a copy must never be (CR 704.5e).
func anywhereButTheStack(g *game.Game, id uuid.UUID) bool {
	if g.Battlefield.Contains(id) || g.Exile.Contains(id) {
		return true
	}
	for _, p := range g.Seats {
		for _, z := range []*game.Zone{p.Hand, p.Graveyard, p.Library, p.Command} {
			if z != nil && z.Contains(id) {
				return true
			}
		}
	}
	return false
}

// twincastBolt casts Lightning Bolt at `victim`, Twincasts it, and
// resolves Twincast keeping the target, leaving the copy on the stack
// above the Bolt. Returns the Bolt and the copy.
func twincastBolt(t *testing.T, g *game.Game, me, victim uuid.UUID) (bolt, cp uuid.UUID) {
	t.Helper()
	bolt = castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	castCatalogSpell(t, g, "Twincast", "Instant", twincastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	p := passUntilPickTarget(t, g, me)
	if p == nil {
		t.Fatal("Twincast opened no re-target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	cp = copyOnStack(g)
	if cp == uuid.Nil {
		t.Fatal("no copy on the stack after Twincast resolved")
	}
	return bolt, cp
}

// TestACounteredTwincastCopyLeavesNoCardBehind — the issue's own
// reproduction, on the cards. The copy is countered, the Bolt resolves:
// 3 damage, not 6, and the graveyard holds the Bolt and Twincast and
// NOT a third, phantom Bolt. The counter still happened, so a
// "whenever a spell is countered" watcher still sees it.
func TestACounteredTwincastCopyLeavesNoCardBehind(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)

	_, cp := twincastBolt(t, g, me, victim)
	counterNow(t, g, cp)

	if anywhereButTheStack(g, cp) {
		t.Fatal("the countered copy landed in a zone — a copy is not a card (CR 707.10a)")
	}
	countered := false
	for _, ev := range g.Events {
		if ev.CardID != cp {
			continue
		}
		if ev.Kind == game.EventCounterSpell {
			countered = true
		}
		if ev.Kind == game.EventZoneMove && ev.NewZone != "" {
			t.Errorf("the copy was announced moving to %q; it goes nowhere", ev.NewZone)
		}
	}
	if !countered {
		t.Error("countering the copy emitted no EventCounterSpell — it was still countered (CR 701.6a)")
	}

	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victim); got != before-3 {
		t.Errorf("life %d → %d, want -3: only the Bolt resolves", before, got)
	}
	if got := graveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d cards, want 2 (Lightning Bolt and Twincast — no phantom copy)", got)
	}
}

// TestABouncedStormCopyDoesNotReachAHand — a Remand-style "return
// target spell to its owner's hand" aimed at a storm copy. The copy
// goes nowhere: the hand does not grow, and the storm spell itself
// still resolves.
func TestABouncedStormCopyDoesNotReachAHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life

	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	p := passUntilPickTarget(t, g, me.ID)
	if p == nil {
		t.Fatal("the storm copy opened no re-target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	cp := copyOnStack(g)
	if cp == uuid.Nil {
		t.Fatal("no storm copy on the stack")
	}
	hand := me.Hand.Size()

	g.WithWriteLock(func() {
		if err := g.ReturnSpellToHandForEffect(cp); err != nil {
			t.Fatalf("ReturnSpellToHandForEffect: %v", err)
		}
	})
	if anywhereButTheStack(g, cp) || g.Stack.Contains(cp) {
		t.Fatal("the bounced copy still exists — it must cease to exist (CR 707.10a)")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d: a copy is not a card and cannot be returned to one", hand, got)
	}

	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 1 {
		t.Errorf("opponent lost %d life, want 1: Grapeshot resolves, its bounced copy does not", got)
	}
}

// TestAvenInterrupterOnACopyExilesAndPlotsNothing — #1318's
// ExileTargetSpell aimed at a Twincast copy. The copy ceases to exist
// instead of being exiled, so "it becomes plotted" has nothing to plot
// and nobody gets a free Lightning Bolt on a later turn.
func TestAvenInterrupterOnACopyExilesAndPlotsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)

	_, cp := twincastBolt(t, g, me, victim)
	aven := castCatalogSpell(t, g, "Aven Interrupter", "Creature — Bird Rogue", avenInterrupterOracle, nil)
	passUntilOnBattlefield(t, g, aven)
	pickCard(t, g, me, cp)
	passPriorityAroundTable(t, g)

	if anywhereButTheStack(g, cp) {
		t.Fatal("the exiled copy is in a zone — it must cease to exist (CR 707.10a)")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(cp); perm.Granted() {
		t.Errorf("the copy was plotted: %+v", perm)
	}
	if got := lifeOf(g, victim); got != before-3 {
		t.Errorf("life %d → %d, want -3: only the Bolt resolves", before, got)
	}
	assertTableMovesOn(t, g)
}
