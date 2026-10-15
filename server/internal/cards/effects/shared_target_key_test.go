package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestSharedTargetKeyIsCardAndAbility — #3817 (ADR 0127, 2026-10-15
// amendment): the owner may answer every copy of a card's trigger with
// one target. The key that groups them is the card and the ability, so
// two Blood Artists share it, and Falkenrath Noble — the same printed
// trigger, the same "target player" clause, another card — does not.
// It rides only the chooser's view.
func TestSharedTargetKeyIsCardAndAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	a2 := pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	noble := pushCatalogPermanent(g, me.ID, "Falkenrath Noble", "Creature — Vampire Noble", falkenrathNobleOracle, false)
	theirs := pushCatalogPermanent(g, opp.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	victim := pushWipeCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}

	keys := map[string]string{} // source → key
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget {
			continue
		}
		k := c.SharedTargetKey()
		if k == "" {
			t.Fatalf("prompt from %s has no shared target key", c.Source)
		}
		keys[c.Source.String()] = k
	}
	if len(keys) != 4 {
		t.Fatalf("got %d target prompts by source, want 4 (two artists, the noble, their artist)", len(keys))
	}
	if keys[a1.String()] != keys[a2.String()] {
		t.Errorf("two Blood Artists: keys %q and %q, want one", keys[a1.String()], keys[a2.String()])
	}
	if keys[noble.String()] == keys[a1.String()] {
		t.Errorf("Falkenrath Noble shares Blood Artist's key %q", keys[a1.String()])
	}
	// The opponent's copy is the same card and ability: the key is the
	// same, and it is the chooser that keeps the two players apart.
	if keys[theirs.String()] != keys[a1.String()] {
		t.Errorf("the opponent's Blood Artist key %q, want the card's %q", keys[theirs.String()], keys[a1.String()])
	}

	raw := protocol.ViewOfGame(g)
	seen := 0
	for _, pc := range protocol.FilterViewFor(raw, me.ID.String()).PendingChoices {
		if pc.Kind != string(game.PendingChoicePickTarget) || pc.Chooser != me.ID.String() {
			continue
		}
		seen++
		if pc.SharedTargetKey != keys[pc.Source] {
			t.Errorf("chooser's view of %s: key %q, want %q", pc.Source, pc.SharedTargetKey, keys[pc.Source])
		}
	}
	if seen != 3 {
		t.Errorf("chooser sees %d of their prompts, want 3", seen)
	}
	for _, pc := range protocol.FilterViewFor(raw, opp.ID.String()).PendingChoices {
		if pc.Chooser == me.ID.String() && pc.SharedTargetKey != "" {
			t.Errorf("another seat sees the chooser's key %q", pc.SharedTargetKey)
		}
	}
}
