package game

import (
	"testing"

	"github.com/google/uuid"
)

// step_entry_layers_test.go — the cleanup step's "until end of turn"
// sweep bumps the layer version, and the cursor walks straight on
// into the next seat's untap and upkeep inside the same write lock.
// Nothing on that path used to recompute, so the untap step and the
// upkeep harvest read the layer cache as it stood BEFORE the effects
// ended: a permanent whose abilities were removed until end of turn
// still looked silenced at the start of the next turn.

// silenceUntilEndOfTurn registers a turn-scoped CR 613.1f "loses all
// abilities" on one permanent and forces the recompute, so the cache
// is known to be the silenced one before the turn ends.
func silenceUntilEndOfTurn(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RegisterTurnScopedStaticForEffect(StaticAbility{
			Layer:            Layer6Ability,
			RemovesAbilities: true,
			AppliesTo:        scopedPinnedTo(id),
			Apply:            func(*Characteristic, *Card, *Game, *Card) {},
		}, uuid.New(), "test — loses all abilities until end of turn")
	})
	if key := CatalogAbilityKey(layeredBattlefieldCard(t, g, id)); key != "" {
		t.Fatalf("setup: the permanent is not silenced (key %q)", key)
	}
}

// endTurnWithoutADiscardPause walks the active seat's turn to its end
// step, which recomputes at every priority boundary on the way, and
// then takes the one advance that runs cleanup and the next seat's
// untap and upkeep in a single chain. The active hand is emptied
// first: a CR 514.1 discard pause would stop the chain at cleanup,
// and the priority boundary after it would recompute and hide the
// read this is about.
func endTurnWithoutADiscardPause(t *testing.T, g *Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; g.Turn.Step != StepEnd; i++ {
		if i > 20 {
			t.Fatalf("never reached the end step (at %s)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { g.Seats[start].Hand.Cards = nil })
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.ActiveSeat == start {
		t.Fatalf("the turn did not end (step %s)", g.Turn.Step)
	}
}

// An untap-step permission whose "loses all abilities until end of
// turn" ended at cleanup works in the very next untap step.
func TestUntapStepReadsLayersRefreshedByTheCleanupSweep(t *testing.T) {
	const oracle = "test-seedborn-probe"
	withCatalogUntapStepPermissions(t, func(id string) []UntapStepPermission {
		if id != oracle {
			return nil
		}
		return []UntapStepPermission{{
			Label: "probe: untap all permanents you control during each other player's untap step",
			AppliesTo: func(_ *Game, source *Card, activePlayer uuid.UUID) bool {
				return activePlayer != source.Controller
			},
			Untaps: func(_ *Game, source, target *Card) bool {
				return target.Controller == source.Controller
			},
		}}
	})

	g := newFourPlayerActiveGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d is active, want 0", g.Turn.ActiveSeat)
	}
	// Seat 2 holds the permission, so seat 1's untap step — the next
	// one — untaps seat 2's permanents only through it.
	holder := g.Seats[2].ID
	probe := pushTappedPermanent(g, holder, "Probe", oracle, "Creature — Spirit", false)
	rock := pushTappedPermanent(g, holder, "Rock", "", "Artifact", true)
	silenceUntilEndOfTurn(t, g, probe)

	endTurnWithoutADiscardPause(t, g)
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("seat %d is active, want 1", g.Turn.ActiveSeat)
	}
	if len(g.TurnScopedStatics) != 0 {
		t.Fatal("the ability loss did not end at cleanup")
	}
	if c, _ := battlefieldCardByID(g, rock); c.Tapped {
		t.Error("the ability loss ended at cleanup, so the permission untaps the rock in the next untap step")
	}
}

// The upkeep that follows in the same chain harvests from fresh
// layers too: "at the beginning of each upkeep" on a permanent whose
// abilities came back at cleanup triggers.
func TestUpkeepHarvestReadsLayersRefreshedByTheCleanupSweep(t *testing.T) {
	const oracle = "test-upkeep-probe"
	var fired []uuid.UUID
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventBeginUpkeep},
			Build: func(ev Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				fired = append(fired, ev.Actor)
				return &StackItem{
					Kind:         StackItemTriggered,
					Controller:   source.Controller,
					Owner:        source.Owner,
					SourceCardID: source.InstanceID,
					Label:        "probe: at the beginning of each upkeep",
				}
			},
		}}
	})

	g := newFourPlayerActiveGame(t)
	probe := pushTappedPermanent(g, g.Seats[2].ID, "Probe", oracle, "Enchantment", false)
	silenceUntilEndOfTurn(t, g, probe)

	endTurnWithoutADiscardPause(t, g)
	next := g.Seats[g.Turn.ActiveSeat].ID
	for _, actor := range fired {
		if actor == next {
			return
		}
	}
	t.Errorf("the ability came back at cleanup, so the next upkeep triggers it (fired for %v)", fired)
}
