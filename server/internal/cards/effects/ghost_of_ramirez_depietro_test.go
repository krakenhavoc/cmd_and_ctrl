package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ghostOfRamirezOracle = "07fe0fb5-cf34-4ba4-a3f3-ac0cc919bf91"

// The evasion: a blocker with toughness 3 or greater is refused, a
// smaller one is not.
func TestGhostOfRamirezCantBeBlockedByToughnessThreeOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghost := b12Push(g, me.ID, "Ghost of Ramirez DePietro", "Legendary Creature — Spirit Pirate", ghostOfRamirezOracle, 2, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 3)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	brAttack(t, g, ghost)

	refused := brRefusal(t, g.DeclareBlocker(wall, ghost), game.BlockReasonCantBeBlockedBy)
	if refused.Label != "creatures with toughness 3 or greater" {
		t.Errorf("refusal label %q", refused.Label)
	}
	if err := g.DeclareBlocker(bear, ghost); err != nil {
		t.Fatalf("a 2-toughness creature may block: %v", err)
	}
}

// The trigger's target: a card discarded or milled this turn, in any
// graveyard — not a creature that died, and not a card that was
// already there. The chosen card goes to its OWNER's hand.
func TestGhostOfRamirezReturnsACardDiscardedOrMilledThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghost := b12Push(g, me.ID, "Ghost of Ramirez DePietro", "Legendary Creature — Spirit Pirate", ghostOfRamirezOracle, 2, 3)

	old := ssCard(opp.Graveyard, opp.ID, "Old Card", "Sorcery", 0, 0)

	opp.Hand.Cards = nil
	discarded := ssCard(opp.Hand, opp.ID, "Discarded Card", "Instant", 0, 0)
	milled := ssCard(me.Library, me.ID, "Milled Card", "Sorcery", 0, 0)
	died := b12Creature(g, opp.ID, "Doomed Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(opp.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
		_ = g.DestroyPermanentForEffect(died)
	})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(discarded) || !me.Graveyard.Contains(milled) || !opp.Graveyard.Contains(died) {
		t.Fatal("setup: the three cards reach their graveyards")
	}

	emitDamageEventForTest(g, ghost, opp.ID, 2, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, discarded) || !hasID(p.PickTargetCards, milled) {
		t.Errorf("a discarded card and a milled card are legal targets: %v", p.PickTargetCards)
	}
	if hasID(p.PickTargetCards, died) {
		t.Error("a creature that died this turn was not discarded or milled")
	}
	if hasID(p.PickTargetCards, old) {
		t.Error("a card already in the graveyard before this turn is not a target")
	}
	if p.PickTargetMin != 0 {
		t.Errorf("up to one target: min %d", p.PickTargetMin)
	}
	pickCard(t, g, me.ID, discarded)
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(discarded) {
		t.Error("the card goes to its owner's hand")
	}
	if me.Hand.Contains(discarded) {
		t.Error("not to the Ghost's controller's hand")
	}
}

// A card discarded this turn that has since left the graveyard and
// come back another way is judged by its latest arrival.
func TestGhostOfRamirezReadsTheLatestArrival(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	card := ssCard(me.Hand, me.ID, "Bear Card", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	var discardedThen bool
	g.ReadSnapshot(func() { discardedThen = discardedOrMilledThisTurn(g, card) })
	if !discardedThen {
		t.Fatal("a card just discarded was discarded this turn")
	}

	// Reanimated, then destroyed: its latest arrival is a death.
	g.WithWriteLock(func() {
		if err := (ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	if !g.Battlefield.Contains(card) {
		t.Fatal("setup: the card is on the battlefield")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(card) })
	passPriorityAroundTable(t, g)
	var stillCounts bool
	g.ReadSnapshot(func() { stillCounts = discardedOrMilledThisTurn(g, card) })
	if stillCounts {
		t.Error("a card whose latest arrival was a death is not 'discarded this turn'")
	}
}
