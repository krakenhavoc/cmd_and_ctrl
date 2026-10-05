package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_cause_discard_sacrifice_test.go — #2178's engine half: what
// the gate does and does not ask. The catalog end-to-end cases live in
// cards/effects/cant_cause_discard_sacrifice_test.go.

const protectedKey = "test/protected"

// withProtection stubs the catalog so a card carrying OracleID
// protectedKey protects its controller from discard and sacrifice.
func withProtection(t *testing.T) {
	t.Helper()
	old := CatalogOpponentEffectProtections
	CatalogOpponentEffectProtections = func(key string) []OpponentEffectProtection {
		if key == protectedKey {
			return []OpponentEffectProtection{{Discard: true, Sacrifice: true}}
		}
		return nil
	}
	t.Cleanup(func() { CatalogOpponentEffectProtections = old })
}

func protectionFixture(t *testing.T) (g *Game, shielded, opponent *Player, handCard, perm uuid.UUID) {
	t.Helper()
	withProtection(t)
	g = newActiveGame(t)
	shielded, opponent = g.Seats[0], g.Seats[1]
	src := NewCard("Shield", shielded.ID)
	src.TypeLine = "Enchantment"
	src.OracleID = protectedKey
	g.Battlefield.PushTop(src)
	h := NewCard("Spare", shielded.ID)
	shielded.Hand.PushTop(h)
	p := NewCard("Fodder", shielded.ID)
	p.TypeLine = "Artifact"
	g.Battlefield.PushTop(p)
	return g, shielded, opponent, h.InstanceID, p.InstanceID
}

func resolvingFor(g *Game, controller uuid.UUID) {
	g.resolving = &resolvingItem{item: &StackItem{Controller: controller}}
}

func TestEffectDiscardFromOpponentIsBlockedButCleanupIsNot(t *testing.T) {
	g, shielded, opponent, hand, _ := protectionFixture(t)
	resolvingFor(g, opponent.ID)

	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.discardCardsLocked(shielded.ID, []uuid.UUID{hand}, discardOptions{cause: DiscardCauseEffect}); err != nil {
		t.Fatal(err)
	}
	if !shielded.Hand.Contains(hand) {
		t.Fatal("an opponent's effect discarded the protected player's card")
	}
	if err := g.discardCardsLocked(shielded.ID, []uuid.UUID{hand}, discardOptions{cause: DiscardCauseCleanup}); err != nil {
		t.Fatal(err)
	}
	if shielded.Hand.Contains(hand) {
		t.Fatal("the cleanup-step discard is a rule, not an opponent's spell")
	}
}

func TestCostDiscardIsNeverBlocked(t *testing.T) {
	g, shielded, opponent, hand, _ := protectionFixture(t)
	resolvingFor(g, opponent.ID)
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.discardCardsLocked(shielded.ID, []uuid.UUID{hand}, discardOptions{cause: DiscardCauseCost}); err != nil {
		t.Fatal(err)
	}
	if shielded.Hand.Contains(hand) {
		t.Fatal("a discard paid as a cost was blocked")
	}
}

func TestOwnEffectDiscardIsNotBlocked(t *testing.T) {
	g, shielded, _, hand, _ := protectionFixture(t)
	resolvingFor(g, shielded.ID)
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.discardCardsLocked(shielded.ID, []uuid.UUID{hand}, discardOptions{cause: DiscardCauseEffect}); err != nil {
		t.Fatal(err)
	}
	if shielded.Hand.Contains(hand) {
		t.Fatal("the player's own spell was stopped from making them discard")
	}
}

func TestSacrificeGateOnlyOnEffectEntryPoints(t *testing.T) {
	g, shielded, opponent, _, perm := protectionFixture(t)
	resolvingFor(g, opponent.ID)
	g.mu.Lock()
	defer g.mu.Unlock()

	// Effect entry points: blocked.
	if err := g.SacrificePermanentForEffect(perm); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(perm) {
		t.Fatal("SacrificePermanentForEffect sacrificed a protected player's permanent")
	}
	if n := g.SacrificeAllForEffect(uuid.Nil, []uuid.UUID{perm}); n != 0 || !g.Battlefield.Contains(perm) {
		t.Fatalf("SacrificeAllForEffect moved a protected permanent (n=%d)", n)
	}
	if g.PlayerSacrificesForEffect(uuid.Nil, shielded.ID, nil, "x") != 0 {
		t.Fatal("a sacrifice prompt was queued for a protected player")
	}

	// Cost path: not blocked, even with a stale opponent item parked in
	// the resolving slot.
	if err := g.sacrificeAnsweredLocked(perm, commanderZoneUnasked); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(perm) {
		t.Fatal("a sacrifice paid as a cost was blocked")
	}
}
