package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func withCatalogUntapStepRestrictions(t *testing.T, fn func(string) []UntapStepRestriction) {
	t.Helper()
	prev := CatalogUntapStepRestrictions
	CatalogUntapStepRestrictions = fn
	t.Cleanup(func() { CatalogUntapStepRestrictions = prev })
}

func cardByIDForUntapTest(g *Game, id uuid.UUID) *Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestUntapRestrictionOnlyAppliesDuringControllersOwnStep(t *testing.T) {
	g := newActiveGame(t)
	active, other := g.Seats[0], g.Seats[1]
	const (
		vault = "restricted-vault"
		muse  = "test-seedborn"
	)
	vaultID := pushTappedPermanent(g, active.ID, "Vault", vault, "Artifact", true)
	pushTappedPermanent(g, active.ID, "Muse", muse, "Creature", false)
	withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
		if key != vault {
			return nil
		}
		return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }}}
	})
	withCatalogUntapStepPermissions(t, func(key string) []UntapStepPermission {
		if key != muse {
			return nil
		}
		return []UntapStepPermission{{AppliesTo: func(_ *Game, source *Card, player uuid.UUID) bool { return source.Controller != player }, Untaps: func(_ *Game, source, target *Card) bool { return target.Controller == active.ID }}}
	})
	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	if !cardByIDForUntapTest(g, vaultID).Tapped {
		t.Fatal("restriction did not hold permanent during its controller's step")
	}
	g.WithWriteLock(func() { g.performUntapStepLocked(other.Seat) })
	if cardByIDForUntapTest(g, vaultID).Tapped {
		t.Fatal("Seedborn-style untap was incorrectly restricted")
	}
}

func TestNextUntapMarkersConsumeAtActualStepAndFollowKey(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	const muse = "marker-seedborn"
	controllerKeyed := pushTappedPermanent(g, a.ID, "Controller keyed", "", "Creature", true)
	playerKeyed := pushTappedPermanent(g, b.ID, "Player keyed", "", "Creature", true)
	pushTappedPermanent(g, b.ID, "Muse", muse, "Creature", false)
	withCatalogUntapStepPermissions(t, func(key string) []UntapStepPermission {
		if key != muse {
			return nil
		}
		return []UntapStepPermission{{AppliesTo: func(_ *Game, source *Card, player uuid.UUID) bool { return source.Controller != player }, Untaps: func(_ *Game, source, target *Card) bool { return target.Controller == source.Controller }}}
	})
	g.WithWriteLock(func() {
		_ = g.SkipNextUntapForEffect(controllerKeyed, uuid.Nil)
		_ = g.SkipNextUntapForEffect(controllerKeyed, uuid.Nil) // dedupe
		_ = g.SkipNextUntapForEffect(playerKeyed, a.ID)
		g.performUntapStepLocked(0)
	})
	if !cardByIDForUntapTest(g, controllerKeyed).Tapped || !cardByIDForUntapTest(g, playerKeyed).Tapped {
		t.Fatal("markers did not hold cards at keyed step")
	}
	if got := len(cardByIDForUntapTest(g, controllerKeyed).NextUntapSkips); got != 0 {
		t.Fatalf("controller marker count = %d, want 0", got)
	}
	if got := len(cardByIDForUntapTest(g, playerKeyed).NextUntapSkips); got != 0 {
		t.Fatalf("player marker count = %d, want 0", got)
	}
	g.WithWriteLock(func() {
		g.performUntapStepLocked(1)
		cardByIDForUntapTest(g, controllerKeyed).Tapped = true
		g.performUntapStepLocked(0)
	})
	if cardByIDForUntapTest(g, controllerKeyed).Tapped || cardByIDForUntapTest(g, playerKeyed).Tapped {
		t.Fatal("consumed markers continued to hold cards")
	}
}

func TestStunPreventsEveryUntapAndRemovesOneWithoutUntapEvent(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushTappedPermanent(g, p.ID, "Stunned", "", "Creature", true)
	g.WithWriteLock(func() { _ = g.applyCounterLocked(id, CounterStun, 2); _ = g.UntapTargetForEffect(id) })
	if c := cardByIDForUntapTest(g, id); !c.Tapped || c.Counters[CounterStun] != 1 {
		t.Fatalf("stun untap = %#v", c)
	}
	beforeUntapEvents := len(g.Events)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id); _ = g.UntapTargetForEffect(id) })
	if c := cardByIDForUntapTest(g, id); c.Tapped || c.Counters[CounterStun] != 0 {
		t.Fatalf("two stun replacements then untap = %#v", c)
	}
	upright := pushTappedPermanent(g, p.ID, "Upright stun", "", "Creature", false)
	g.WithWriteLock(func() { _ = g.applyCounterLocked(upright, CounterStun, 1); _ = g.UntapTargetForEffect(upright) })
	if got := cardByIDForUntapTest(g, upright).Counters[CounterStun]; got != 1 {
		t.Fatalf("untapped permanent spent stun: %d", got)
	}
	for _, ev := range g.Events[:beforeUntapEvents] {
		if ev.Kind == EventUntapCard && ev.CardID == id {
			t.Fatal("stun replacement emitted untap")
		}
	}
}

func TestStepRestrictionKeepsStunButSandboxUntapConsumesItAndKeepsMarker(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	const frozen = "stun-restricted"
	id := pushTappedPermanent(g, p.ID, "Frozen", frozen, "Artifact", true)
	withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
		if key == frozen {
			return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }}}
		}
		return nil
	})
	g.WithWriteLock(func() {
		_ = g.applyCounterLocked(id, CounterStun, 1)
		_ = g.SkipNextUntapForEffect(id, p.ID)
		g.performUntapStepLocked(0)
	})
	if c := cardByIDForUntapTest(g, id); !c.Tapped || c.Counters[CounterStun] != 1 {
		t.Fatalf("held step spent stun: %#v", c)
	}
	// A fresh marker demonstrates that the sandbox action ignores markers,
	// but it still routes through stun's common untap primitive.
	g.WithWriteLock(func() { _ = g.SkipNextUntapForEffect(id, p.ID); g.untapAllForLocked(0) })
	if c := cardByIDForUntapTest(g, id); !c.Tapped || c.Counters[CounterStun] != 0 || len(c.NextUntapSkips) != 1 {
		t.Fatalf("sandbox stun/marker result = %#v", c)
	}
}

func TestNextUntapMarkerCloneSnapshotAndZoneReset(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() { _ = g.SkipNextUntapForEffect(id, p.ID) })
	clone := g.Clone()
	if got := clone.Battlefield.Cards[0].NextUntapSkips; len(got) != 1 || got[0].Player != p.ID {
		t.Fatalf("clone markers = %#v", got)
	}
	g.WithWriteLock(func() { _ = g.SkipNextUntapForEffect(id, uuid.Nil) })
	if got := clone.Battlefield.Cards[0].NextUntapSkips; len(got) != 1 {
		t.Fatalf("clone marker slice aliased live card: %#v", got)
	}
	g.WithWriteLock(func() { g.RestoreFrom(clone) })
	if got := cardByIDForUntapTest(g, id).NextUntapSkips; len(got) != 1 || got[0].Player != p.ID {
		t.Fatalf("RestoreFrom markers = %#v", got)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.Battlefield.Cards[0].NextUntapSkips; len(got) != 1 || got[0].Player != p.ID {
		t.Fatalf("restored markers = %#v", got)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: p.ID}, id); err != nil {
		t.Fatal(err)
	}
	if got := g.Seats[0].Graveyard.Cards[0].NextUntapSkips; len(got) != 0 {
		t.Fatalf("zone-change marker = %#v", got)
	}
}

func TestUntapMarkersFollowControlAndConsumeWhileUntapped(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	id := pushTappedPermanent(g, a.ID, "Stolen", "", "Creature", true)
	upright := pushTappedPermanent(g, a.ID, "Upright", "", "Creature", false)
	g.WithWriteLock(func() {
		_ = g.SkipNextUntapForEffect(id, uuid.Nil)
		_ = g.SkipNextUntapForEffect(upright, a.ID)
		cardByIDForUntapTest(g, id).Controller = b.ID
		// The skipped/stasis branch never calls performUntapStepLocked, so
		// both markers remain until an actual step happens.
		if len(cardByIDForUntapTest(g, id).NextUntapSkips) != 1 {
			t.Fatal("marker changed before a step")
		}
		g.performUntapStepLocked(1)
	})
	if !cardByIDForUntapTest(g, id).Tapped || len(cardByIDForUntapTest(g, id).NextUntapSkips) != 0 {
		t.Fatal("controller-keyed marker did not follow control")
	}
	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	if got := cardByIDForUntapTest(g, upright).NextUntapSkips; len(got) != 0 {
		t.Fatalf("upright marker did not consume: %#v", got)
	}
}

func TestSkippedUntapStepDoesNotConsumeMarker(t *testing.T) {
	// This drives the real step-transition replacement path rather than
	// merely omitting performUntapStepLocked: Stasis-style cancellation must
	// leave a "next untap step" marker for the next step that actually runs.
	g := newGameWithStepReplacements(t, skipStepForSeat(StepUntap, 1, true, "skip untap"))
	p := g.Seats[1]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() { _ = g.SkipNextUntapForEffect(id, p.ID) })
	advanceUntil(t, g, 20, func() bool { return g.Turn.ActiveSeat == 1 && g.Turn.Step == StepUpkeep })
	if c := cardByIDForUntapTest(g, id); !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Fatalf("skipped step consumed marker: %#v", c)
	}
}

func TestPausedUntapTransitionDoesNotConsumeMarkerBeforeResolution(t *testing.T) {
	g := newGameWithStepReplacements(t,
		watchStepForSeat(StepUntap, 1, "watch untap"),
		skipStepForSeat(StepUntap, 1, false, "skip untap"),
	)
	p := g.Seats[1]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() { _ = g.SkipNextUntapForEffect(id, p.ID) })
	advanceUntil(t, g, 20, func() bool { return len(g.PendingChoices) > 0 })
	if got := cardByIDForUntapTest(g, id).NextUntapSkips; len(got) != 1 {
		t.Fatalf("paused transition consumed marker: %#v", got)
	}
	prompt := g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatal(err)
	}
	if got := cardByIDForUntapTest(g, id).NextUntapSkips; len(got) != 1 {
		t.Fatalf("cancelled resumed transition consumed marker: %#v", got)
	}
}

func TestUntapRestrictionReadsPowerAndAbilityRemoval(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	const meekstone = "power-restriction"
	target := pushTappedPermanent(g, p.ID, "Big", "", "Creature", true)
	source := pushTappedPermanent(g, p.ID, "Meek", meekstone, "Artifact", false)
	cardByIDForUntapTest(g, target).Power = 3
	withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
		if key != meekstone {
			return nil
		}
		return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, _ *Card) bool { return target.Effective().Power >= 3 }}}
	})
	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	if !cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("power restriction did not apply")
	}
	g.WithWriteLock(func() {
		e := cardByIDForUntapTest(g, source).Effective()
		e.AbilitiesRemoved = true
		cardByIDForUntapTest(g, source).effective = &e
		g.performUntapStepLocked(0)
	})
	if cardByIDForUntapTest(g, target).Tapped {
		t.Fatal("ability removal did not disable restriction")
	}
}

func TestAutoTapSpendsFrozenSourceLastButCanFallBack(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	const frozen = "frozen-mountain"
	frozenID := pushBattlefieldForTest(g, p.ID, "Frozen Mountain", "Basic Land — Mountain", frozen)
	normalID := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
		if key != frozen {
			return nil
		}
		return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }}}
	})
	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0)
	if !ok || len(plan) != 1 || plan[0] != normalID {
		t.Fatalf("normal source was not preferred: %v", plan)
	}
	cardByIDForUntapTest(g, normalID).Tapped = true
	plan, ok = g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0)
	if !ok || len(plan) != 1 || plan[0] != frozenID {
		t.Fatalf("frozen fallback was excluded: %v", plan)
	}
}

func TestAutoTapReservesFrozenColorlessSourcesForGenericFallback(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	const vault = "frozen-colorless-vault"
	withCatalogHook(t, func(key string) []ManaAbilityShape {
		if key == vault {
			return []ManaAbilityShape{{TapCost: true, Produced: "{C}"}}
		}
		return nil
	})
	frozenID := pushBattlefieldForTest(g, p.ID, "Frozen Vault", "Artifact", vault)
	normalID := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
		if key != vault {
			return nil
		}
		return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, source *Card) bool {
			return target.InstanceID == source.InstanceID
		}}}
	})

	// Generic recruitment must reserve the frozen colorless source even
	// though its ordinary generic preference is ahead of a Mountain.
	plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 || plan[0] != normalID {
		t.Fatalf("generic plan spent frozen Vault before Mountain: %v", plan)
	}
	// Colored demand continues to use the ordinary colored source.
	plan, ok = g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0)
	if !ok || len(plan) != 1 || plan[0] != normalID {
		t.Fatalf("colored plan = %v, want normal Mountain", plan)
	}
	// Once no ordinary source remains, the frozen source is still a
	// legal generic fallback rather than an excluded source.
	cardByIDForUntapTest(g, normalID).Tapped = true
	plan, ok = g.AutoTapForCost(p.ID, costFor(t, "{1}"), 0)
	if !ok || len(plan) != 1 || plan[0] != frozenID {
		t.Fatalf("generic fallback = %v, want frozen Vault", plan)
	}
}

func TestPublicAutoTapRefreshesStaleUntapRestrictionCharacteristics(t *testing.T) {
	t.Run("Meekstone power", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		const meekstone = "autotap-meekstone"
		withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
			if key != meekstone {
				return nil
			}
			return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, _ *Card) bool {
				return target.Effective().Power >= 3
			}}}
		})
		pushBattlefieldForTest(g, p.ID, "Meekstone", "Artifact", meekstone)
		bigID := pushBattlefieldForTest(g, p.ID, "Animated Mountain", "Basic Land — Mountain", "")
		normalID := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
		big := cardByIDForUntapTest(g, bigID)
		big.Power = 3
		stale := big.printedCharacteristic()
		stale.Power = 2
		big.effective = &stale
		g.layerVersion.Add(1)

		plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0)
		if !ok || len(plan) != 1 || plan[0] != normalID {
			t.Fatalf("stale-power plan = %v, want unrestricted Mountain", plan)
		}
	})

	t.Run("silenced self restriction", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		const selfRestricted = "autotap-self-restricted"
		withCatalogUntapStepRestrictions(t, func(key string) []UntapStepRestriction {
			if key != selfRestricted {
				return nil
			}
			return []UntapStepRestriction{{Restricts: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			}}}
		})
		frozenID := pushBattlefieldForTest(g, p.ID, "Restricted Mountain", "Basic Land — Mountain", selfRestricted)
		normalID := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
		frozen := cardByIDForUntapTest(g, frozenID)
		stale := frozen.printedCharacteristic()
		stale.AbilitiesRemoved = true
		frozen.effective = &stale
		g.layerVersion.Add(1)

		plan, ok := g.AutoTapForCost(p.ID, costFor(t, "{R}"), 0)
		if !ok || len(plan) != 1 || plan[0] != normalID {
			t.Fatalf("stale-silence plan = %v, want unrestricted Mountain", plan)
		}
	})
}
