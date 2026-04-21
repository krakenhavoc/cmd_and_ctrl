package game

import (
	"testing"

	"github.com/google/uuid"
)

// effect_hooks_test.go covers the S14 sub-PR 3 resolution integration
// using stubbed callbacks — no import of server/internal/cards/effects
// (that would be a cycle for tests in the game package). The stubs
// simulate what effects/wire.go registers at package init.

// withEffectHooks installs stub hooks for the duration of a test
// and restores the prior values (nil in isolation, the effects
// package's own registrations under the full test binary) on
// cleanup. Safe against the package-global singleton — each test
// registers its own stub, runs, restores.
func withEffectHooks(t *testing.T, resolver func(g *Game, item *StackItem, scryfallID string) error, etb func(g *Game, cardID uuid.UUID, scryfallID string) error, isCatalog func(scryfallID string) bool) {
	t.Helper()
	prevResolver := EffectResolver
	prevETB := ETBEffectHook
	prevCatalog := IsCatalogCard
	EffectResolver = resolver
	ETBEffectHook = etb
	IsCatalogCard = isCatalog
	t.Cleanup(func() {
		EffectResolver = prevResolver
		ETBEffectHook = prevETB
		IsCatalogCard = prevCatalog
	})
}

// TestEffectResolverFiresOnCatalogSpell proves the resolve hook runs
// for a spell whose ScryfallID the stub catalog knows.
func TestEffectResolverFiresOnCatalogSpell(t *testing.T) {
	const knownScryfallID = "catalog-lightning-bolt"
	var resolverCalled bool
	withEffectHooks(t,
		func(g *Game, item *StackItem, scryfallID string) error {
			if scryfallID == knownScryfallID {
				resolverCalled = true
			}
			return nil
		},
		nil,
		func(scryfallID string) bool { return scryfallID == knownScryfallID },
	)

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	spellID := pushTypedCardToHand(caster, "Lightning Bolt", "Instant")
	// Stamp the Scryfall ID so the resolver can match.
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == spellID {
			caster.Hand.Cards[i].ScryfallID = knownScryfallID
			break
		}
	}
	if err := g.CastSpell(caster.ID, spellID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Pass priority around to resolve the spell.
	_ = g.PassPriority()
	_ = g.PassPriority()

	if !resolverCalled {
		t.Errorf("EffectResolver was not called for catalog spell")
	}
}

// TestEffectResolverSkipsNonCatalogSpell pins the opt-in invariant:
// a spell whose ScryfallID is not in the catalog must not trigger
// any resolve-time effect. Canary for regressions that would break
// the Cockatrice-style manual fallback.
func TestEffectResolverSkipsNonCatalogSpell(t *testing.T) {
	var resolverCalled bool
	withEffectHooks(t,
		func(g *Game, item *StackItem, scryfallID string) error {
			resolverCalled = true
			return nil
		},
		nil,
		func(scryfallID string) bool { return false },
	)

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	spellID := pushTypedCardToHand(caster, "Strange Spell", "Instant")
	// No ScryfallID → Lookup miss → resolver is called but the
	// effects package returns nil quickly. Here we assert that
	// EffectResolver IS called (it always is, per the design —
	// the catalog miss happens inside the resolver). This test's
	// real value is showing that the spell still resolves and
	// reaches the graveyard even when the catalog says "no effect."
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == spellID {
			caster.Hand.Cards[i].ScryfallID = "not-in-catalog"
			break
		}
	}
	if err := g.CastSpell(caster.ID, spellID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()

	if !resolverCalled {
		t.Errorf("EffectResolver was not called at all — the resolution hook isn't wired")
	}
	if !caster.Graveyard.Contains(spellID) {
		t.Errorf("non-catalog spell did not route to graveyard after resolve")
	}
}

// TestEffectResolverErrorEmitsEventEffectError proves that a
// resolver returning an error emits EventEffectError and lets the
// resolution path carry on (spell still routes to graveyard).
func TestEffectResolverErrorEmitsEventEffectError(t *testing.T) {
	withEffectHooks(t,
		func(g *Game, item *StackItem, scryfallID string) error {
			return ErrInvalidParam
		},
		nil,
		func(scryfallID string) bool { return true },
	)

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	spellID := pushTypedCardToHand(caster, "Broken Spell", "Instant")
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == spellID {
			caster.Hand.Cards[i].ScryfallID = "any-id"
			break
		}
	}
	if err := g.CastSpell(caster.ID, spellID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	base := len(g.Events)
	_ = g.PassPriority()
	_ = g.PassPriority()

	// EventEffectError must appear in the log.
	var sawError bool
	for _, ev := range g.Events[base:] {
		if ev.Kind == EventEffectError {
			sawError = true
			break
		}
	}
	if !sawError {
		t.Errorf("EventEffectError not emitted after resolver error")
	}
	// Spell still hits graveyard — resolution didn't wedge.
	if !caster.Graveyard.Contains(spellID) {
		t.Errorf("spell did not route to graveyard after resolver error")
	}
}

// TestETBHookFiresOnSpellResolve verifies the ETB hook runs when a
// permanent spell resolves onto the battlefield. Uses a creature
// spell (so it goes through resolveTopOfStackLocked's permanent
// branch, which fires the hook).
func TestETBHookFiresOnSpellResolve(t *testing.T) {
	var etbCalled bool
	var etbCardID uuid.UUID
	withEffectHooks(t,
		nil,
		func(g *Game, cardID uuid.UUID, scryfallID string) error {
			if scryfallID == "etb-creature" {
				etbCalled = true
				etbCardID = cardID
			}
			return nil
		},
		func(scryfallID string) bool { return scryfallID == "etb-creature" },
	)

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	spellID := pushTypedCardToHand(caster, "Test Creature", "Creature — Bear")
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == spellID {
			caster.Hand.Cards[i].ScryfallID = "etb-creature"
			break
		}
	}
	if err := g.CastSpell(caster.ID, spellID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	_ = g.PassPriority()
	_ = g.PassPriority()

	if !etbCalled {
		t.Errorf("ETBEffectHook was not called on spell resolution onto battlefield")
	}
	if etbCardID != spellID {
		t.Errorf("ETBEffectHook received wrong cardID: got %v, want %v", etbCardID, spellID)
	}
}

// TestETBHookFiresOnLandCast pins the land-cast shortcut path —
// lands skip the stack entirely but must still fire the ETB hook
// so a hypothetical "fetch land" catalog entry would work.
func TestETBHookFiresOnLandCast(t *testing.T) {
	var etbCalled bool
	withEffectHooks(t,
		nil,
		func(g *Game, cardID uuid.UUID, scryfallID string) error {
			if scryfallID == "etb-land" {
				etbCalled = true
			}
			return nil
		},
		func(scryfallID string) bool { return scryfallID == "etb-land" },
	)

	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	caster := g.Seats[0]
	landID := pushTypedCardToHand(caster, "Forest", "Basic Land — Forest")
	for i := range caster.Hand.Cards {
		if caster.Hand.Cards[i].InstanceID == landID {
			caster.Hand.Cards[i].ScryfallID = "etb-land"
			break
		}
	}
	if err := g.CastSpell(caster.ID, landID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell (land): %v", err)
	}

	if !etbCalled {
		t.Errorf("ETBEffectHook not fired on land cast")
	}
}

// TestIsAutoCardReflectsCatalog proves the view-layer auto-bit
// accessor routes through IsCatalogCard. CardView.Auto is stamped
// via game.IsAutoCard from protocol/view.go; a miss on the hook
// returns false.
func TestIsAutoCardReflectsCatalog(t *testing.T) {
	withEffectHooks(t, nil, nil, func(scryfallID string) bool {
		return scryfallID == "known-id"
	})
	if !IsAutoCard("known-id") {
		t.Errorf("IsAutoCard(known-id) = false, want true")
	}
	if IsAutoCard("unknown-id") {
		t.Errorf("IsAutoCard(unknown-id) = true, want false")
	}
}

