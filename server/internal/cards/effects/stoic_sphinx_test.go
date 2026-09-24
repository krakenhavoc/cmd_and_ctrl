package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stoic_sphinx_test.go — #1325: the card proof for the last quarter
// of layer invalidation on non-battlefield inputs (docs/engine-seams.md,
// "Layer cache invalidation on hand, life, attack and graveyard
// events" — the spells-cast count was the one #1117's own note
// flagged as still open). No other event happens between the cast and
// the read, which is the property that was missing: a stale cache
// survives fine until something unrelated invalidates it, and that is
// exactly what hid this gap.

const stoicSphinxOracle = "d924c13e-f266-47b0-9711-ba9a9b9df1cb"

// pushStoicSphinx seeds a Stoic Sphinx through the layer-aware helper
// (pushBattlefieldCardWithTimestamp) rather than pushCatalogPermanent
// — this card's whole point is a Layer 6 static gated on a
// non-battlefield input, and only the layer-aware push reliably bumps
// layerVersion so the first read forces a real recompute instead of
// falling back to the printed characteristic.
func pushStoicSphinx(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Stoic Sphinx",
		TypeLine:   "Creature — Sphinx",
		OracleID:   stoicSphinxOracle,
		Power:      5,
		Toughness:  3,
		Owner:      owner,
		Controller: owner,
	})
}

// TestStoicSphinxHasHexproofUntilItsControllerCastsASpell is the
// end-to-end shape: hexproof reads live before any spell is cast this
// turn, and turns off the instant one is — the ACTUAL resolved
// characteristic, not merely a cache that eventually catches up.
func TestStoicSphinxHasHexproofUntilItsControllerCastsASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sphinx := pushStoicSphinx(g, me.ID)

	if got := effectiveAbilities(t, g, sphinx); !containsString(got, "hexproof") {
		t.Fatalf("expected hexproof before any spell was cast this turn, got %v", got)
	}

	castCatalogSpell(t, g, "Nothing Much", "Instant", "stoic-sphinx-test-filler-spell", nil)

	if got := effectiveAbilities(t, g, sphinx); containsString(got, "hexproof") {
		t.Fatalf("expected hexproof to be gone once its controller cast a spell this turn, got %v", got)
	}
}

// TestStoicSphinxPrintsFlashAndFlying pins the two combat keywords —
// PrintedKeywords, not the conditional grant above.
func TestStoicSphinxPrintsFlashAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sphinx := pushStoicSphinx(g, me.ID)

	got := effectiveAbilities(t, g, sphinx)
	if !containsString(got, "flash") || !containsString(got, "flying") {
		t.Errorf("abilities = %v, want flash and flying", got)
	}
}

// TestStoicSphinxHexproofIsPerController: it reads OWN spells cast,
// not the table's — an opponent casting a spell must not strip it.
func TestStoicSphinxHexproofIsPerController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sphinx := pushStoicSphinx(g, me.ID)

	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{opp.ID: {Total: 1}}
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID})
	})

	if got := effectiveAbilities(t, g, sphinx); !containsString(got, "hexproof") {
		t.Errorf("an opponent's cast stripped hexproof: %v", got)
	}
}
