package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_scoped_statics_test.go pins the S32 registry's mechanics at
// the engine level: the layer engine sees floating effects, CR
// 613.7 timestamp ordering holds between a floating effect and a
// battlefield one, the cleanup sweep runs at the right moment, and
// the undo stack carries the list. Card-level behaviour (Giant
// Growth, Overrun, Aang) lives in the effects package.

// pushScopedTestCreature seeds a creature on the battlefield and
// fires the zone-move event so the layer listener stamps
// EnteredBattlefieldAt — the stamp the affected-set predicates key
// on.
func pushScopedTestCreature(g *Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  id,
			OldZone: ZoneHand,
			NewZone: ZoneBattlefield,
		})
	})
	return id
}

// scopedEffectivePT forces a recompute and reads the card's
// post-layer power and toughness.
func scopedEffectivePT(t *testing.T, g *Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				p, tough, found = eff.Power, eff.Toughness, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return p, tough
}

// scopedPinnedTo builds the "this one permanent" AppliesTo predicate the
// card-facing primitives build, without importing the effects
// package (which would be an import cycle).
func scopedPinnedTo(id uuid.UUID) func(*Card, *Game, *Card) bool {
	return func(target *Card, _ *Game, _ *Card) bool {
		return target.InstanceID == id
	}
}

// TestScopedStaticAppliesImmediately is the headline: a
// floating +3/+3 registered with no permanent behind it is visible
// on the very next recompute.
func TestScopedStaticAppliesImmediately(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)

	if p, tough := scopedEffectivePT(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("baseline P/T = %d/%d, want 2/2", p, tough)
	}

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power += 3
				c.Toughness += 3
			},
		}, uuid.New(), "test — +3/+3", g.UntilEndOfTurnDuration())
	})

	if p, tough := scopedEffectivePT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("pumped P/T = %d/%d, want 5/5", p, tough)
	}
}

// TestScopedStaticExpiresAtCleanup walks the cursor through the
// cleanup step and checks the grant is gone — both from the
// registry and from the card's effective characteristic.
func TestScopedStaticExpiresAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power += 3
				c.Toughness += 3
			},
		}, uuid.New(), "test — +3/+3", g.UntilEndOfTurnDuration())
	})
	if p, _ := scopedEffectivePT(t, g, bear); p != 5 {
		t.Fatalf("setup: power = %d, want 5", p)
	}

	advancePastScopedCleanup(t, g)

	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("ScopedStatics = %d after cleanup, want 0", n)
	}
	if p, tough := scopedEffectivePT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("post-cleanup P/T = %d/%d, want 2/2", p, tough)
	}
}

// TestScopedStaticFromEndStepExpiresSameTurn is the CR 514.2
// subtlety the ExpiresAfterTurn field exists for: "until end of
// turn" created DURING the end step still ends at this turn's
// cleanup. A duration that meant "survive until the next cleanup I
// haven't seen yet" would leak the grant into the following turn.
func TestScopedStaticFromEndStepExpiresSameTurn(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)

	// Walk to the end step of the current turn.
	for i := 0; i < 20 && g.Turn.Step != StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != StepEnd {
		t.Fatalf("never reached the end step (at %q)", g.Turn.Step)
	}
	seatAtGrant := g.Turn.ActiveSeat
	turnsBegunAtGrant := g.Seats[seatAtGrant].TurnsBegun

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power += 3
			},
		}, uuid.New(), "end-step grant", g.UntilEndOfTurnDuration())
	})
	if p, _ := scopedEffectivePT(t, g, bear); p != 5 {
		t.Fatalf("setup: power = %d, want 5", p)
	}
	if got := g.ScopedStatics[0].Duration; got.Kind != UntilEndOfTurn ||
		got.ExpiresAfterTurnsBegun != turnsBegunAtGrant {
		t.Fatalf("duration = %+v, want until end of turn stamped on the turn it was created in (%d)",
			got, turnsBegunAtGrant)
	}

	// One step: end step → cleanup, which sweeps and auto-advances
	// to the next seat's untap. (Turn.Number counts rounds, so the
	// seat is what changes here, not the number.)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.ActiveSeat == seatAtGrant {
		t.Fatalf("cursor did not leave seat %d's turn (at %q)", seatAtGrant, g.Turn.Step)
	}
	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("grant made in the end step survived its own turn's cleanup (%d entries)", n)
	}
	if p, _ := scopedEffectivePT(t, g, bear); p != 2 {
		t.Errorf("post-cleanup power = %d, want 2", p)
	}
}

// TestScopedStaticComposesWithAnthemInLayerOrder is the CR
// 613.7 / 613.1 pin: a floating layer-7b "base P/T becomes 1/1"
// applies BEFORE a layer-7c +1/+1 regardless of which was created
// first, so the 4/4 ends up 2/2 and not 1/1. Sub-layer beats
// timestamp — that is the whole point of having layers.
func TestScopedStaticComposesWithAnthemInLayerOrder(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 4, 4)

	// The 7c modifier is created FIRST (earliest timestamp).
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power++
				c.Toughness++
			},
		}, uuid.New(), "anthem-shaped +1/+1", g.UntilEndOfTurnDuration())
	})
	// The 7b setter is created SECOND (latest timestamp) and must
	// still apply first.
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7B_Set,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power, c.Toughness = 1, 1
			},
		}, uuid.New(), "base P/T 1/1", g.UntilEndOfTurnDuration())
	})

	if p, tough := scopedEffectivePT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 2/2 (7b sets 1/1, then 7c adds +1/+1)", p, tough)
	}
}

// TestScopedStaticsSortByTimestampWithinALayer is the CR 613.7
// pin proper: two effects in the SAME sub-layer, where order is
// observable because the later one overwrites the earlier.
func TestScopedStaticsSortByTimestampWithinALayer(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)

	setPT := func(p, tough int, label string) {
		g.WithWriteLock(func() {
			g.RegisterScopedStaticForEffect(StaticAbility{
				Layer:     Layer7PT,
				SubLayer:  SubLayer7B_Set,
				AppliesTo: scopedPinnedTo(bear),
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Power, c.Toughness = p, tough
				},
			}, uuid.New(), label, g.UntilEndOfTurnDuration())
		})
	}
	setPT(1, 1, "base 1/1")
	setPT(5, 5, "base 5/5")

	if p, tough := scopedEffectivePT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("P/T = %d/%d, want 5/5 (the later timestamp wins in 7b)", p, tough)
	}
}

// TestScopedStaticSurvivesItsSourceLeaving is the reason the
// registry exists at all: the effect must outlive the card that
// made it. A battlefield static would vanish with its source.
func TestScopedStaticSurvivesItsSourceLeaving(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0].ID
	bear := pushScopedTestCreature(g, owner, 2, 2)
	sourcePermanent := pushScopedTestCreature(g, owner, 1, 1)

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power += 3
			},
		}, sourcePermanent, "grant from a permanent that is about to die", g.UntilEndOfTurnDuration())
	})

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: owner},
		sourcePermanent,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	if p, _ := scopedEffectivePT(t, g, bear); p != 5 {
		t.Errorf("power = %d, want 5 — the grant must outlive its source", p)
	}
	if got := g.ScopedStatics[0].Source.InstanceID; got != sourcePermanent {
		t.Errorf("stored source LKI = %s, want %s", got, sourcePermanent)
	}
}

// TestCloneCopiesScopedStatics — an undo snapshot taken while a
// grant is live carries it, and the cleanup sweep on the original
// must not reach into the snapshot's copy. The sweep allocates a
// fresh slice for exactly this reason; an in-place `s[:0]`
// compaction would rewrite the snapshot's backing array.
func TestCloneCopiesScopedStatics(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer6Ability,
			AppliesTo: func(*Card, *Game, *Card) bool { return false },
			Apply:     func(*Characteristic, *Card, *Game, *Card) {},
		}, uuid.New(), "grant", g.UntilEndOfTurnDuration())
	})

	snap := g.Clone()
	if len(snap.ScopedStatics) != 1 {
		t.Fatalf("clone dropped the live scoped static")
	}

	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if len(g.ScopedStatics) != 0 {
		t.Fatalf("sweep left %d entries on the original", len(g.ScopedStatics))
	}
	if len(snap.ScopedStatics) != 1 {
		t.Error("sweeping the original reached the clone (aliased slice)")
	}
	if snap.ScopedStatics[0].Label != "grant" {
		t.Errorf("clone's entry = %q, want %q", snap.ScopedStatics[0].Label, "grant")
	}
}

// TestRestoreFromRollsBackScopedStatics is the undo direction:
// a grant registered after the snapshot must be gone once the
// snapshot is restored, and the restored board must recompute
// without it.
func TestRestoreFromRollsBackScopedStatics(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	snap := g.Clone()

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: scopedPinnedTo(bear),
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power += 3
				c.Toughness += 3
			},
		}, uuid.New(), "Giant Growth — +3/+3", g.UntilEndOfTurnDuration())
	})
	if p, _ := scopedEffectivePT(t, g, bear); p != 5 {
		t.Fatalf("setup: power = %d, want 5", p)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("RestoreFrom kept %d scoped statics, want 0", n)
	}
	if p, tough := scopedEffectivePT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("post-undo P/T = %d/%d, want 2/2", p, tough)
	}
}

// TestClearExpiredScopedStaticsIsIdempotent — the cleanup hook
// runs a second time after a discard pause drains, so the sweep
// must tolerate being called with nothing left to do without
// churning the layer version.
func TestClearExpiredScopedStaticsIsIdempotent(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer6Ability,
			AppliesTo: func(*Card, *Game, *Card) bool { return false },
			Apply:     func(*Characteristic, *Card, *Game, *Card) {},
		}, uuid.New(), "grant", g.UntilEndOfTurnDuration())
	})

	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	versionAfterFirst := g.layerVersion.Load()
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })

	if n := len(g.ScopedStatics); n != 0 {
		t.Errorf("second sweep left %d entries", n)
	}
	if got := g.layerVersion.Load(); got != versionAfterFirst {
		t.Errorf("no-op sweep bumped the layer version (%d → %d)", versionAfterFirst, got)
	}
}

// advancePastScopedCleanup walks the cursor until the active seat changes,
// which means this seat's cleanup step ran. (Turn.Number counts
// rounds, not seat-turns, so the seat is the reliable marker.)
func advancePastScopedCleanup(t *testing.T, g *Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 30; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}
