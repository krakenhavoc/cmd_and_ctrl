package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// any_player_activation_test.go — ADR 0106 §1 (#1793): "Any player may
// activate this ability" (CR 602.2, CR 602.1b). The enumerator's and the
// view's halves are pinned in their own packages; this pins the
// activation path and the facts the ADR says already read the activator.

const anyPlayerOracle = "test-any-player-hourglass"

// stubAnyPlayerCard is one permanent with two rows: index 0 is its
// controller's alone, index 1 is an any-player row whose effect records
// who "you" was at resolution.
func stubAnyPlayerCard(t *testing.T, resolvedFor *uuid.UUID) {
	t.Helper()
	stubActivatedFor(t, anyPlayerOracle,
		ActivatedAbilityShape{
			Label:  "{1}: Controller-only.",
			Cost:   AbilityCost{Mana: "{1}"},
			Effect: func(*Game, *StackItem) error { return nil },
		},
		ActivatedAbilityShape{
			Label:     "{1}: Anyone. Any player may activate this ability.",
			Cost:      AbilityCost{Mana: "{1}"},
			AnyPlayer: true,
			Purpose:   Purpose{Draws: 1},
			Condition: func(g *Game, controller, _ uuid.UUID) bool {
				// "Activate only during your turn", where "your" is the
				// activator's (CR 109.5): read with the ACTIVATOR.
				return g.Seats[g.Turn.ActiveSeat].ID == controller
			},
			Effect: func(_ *Game, item *StackItem) error {
				*resolvedFor = item.Controller
				return nil
			},
		},
	)
}

func anyPlayerItemFor(g *Game, source uuid.UUID) *StackItem {
	for _, it := range g.StackMeta {
		if it.SourceCardID == source && it.Kind == StackItemActivated {
			return it
		}
	}
	return nil
}

// TestAnyPlayerRowIsAnyonesToActivate is the headline: a non-controller
// may activate the any-player row and only that row, with their own
// mana, and the ability on the stack is theirs (CR 602.2, 602.1a,
// 113.8).
func TestAnyPlayerRowIsAnyonesToActivate(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	active, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var resolvedFor uuid.UUID
	stubAnyPlayerCard(t, &resolvedFor)
	src := pushGateCard(g, "Test Hourglass", "Artifact", anyPlayerOracle, other.ID)

	// The controller-only row stays its controller's (CR 602.2).
	active.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(active.ID, src, 0, ActivateAbilityParams{Strict: true}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("non-controller on the controller-only row: err = %v, want ErrCardCallerMismatch", err)
	}
	if len(active.ManaPool) != 2 {
		t.Fatalf("a refused activation spent mana")
	}
	if err := g.ActivateCatalogAbility(active.ID, src, 1, ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("non-controller on the any-player row: %v", err)
	}
	if len(active.ManaPool) != 1 {
		t.Errorf("activator's pool = %d after paying {1}, want 1 (CR 602.1a)", len(active.ManaPool))
	}
	it := anyPlayerItemFor(g, src)
	if it == nil {
		t.Fatal("no ability on the stack")
	}
	if it.Controller != active.ID {
		t.Errorf("item controller = %v, want the activator (CR 113.8)", it.Controller)
	}
	if bc := it.BaseController; bc != uuid.Nil && bc != active.ID {
		t.Errorf("item base controller = %v, want the activator", bc)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	for i := 0; anyPlayerItemFor(g, src) != nil && i < 8; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if resolvedFor != active.ID {
		t.Errorf(`"you" at resolution = %v, want the activator %v (CR 109.5)`, resolvedFor, active.ID)
	}
}

// TestAnyPlayerConditionReadsTheActivator pins the ADR's claim that the
// Condition is asked with the ACTIVATOR as "you": the controller, off
// turn, may not activate a row whose condition says "during your turn",
// while the active player — who does not control it — may.
func TestAnyPlayerConditionReadsTheActivator(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	active, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var resolvedFor uuid.UUID
	stubAnyPlayerCard(t, &resolvedFor)
	src := pushGateCard(g, "Test Hourglass", "Artifact", anyPlayerOracle, other.ID)
	if err := g.ActivateCatalogAbility(other.ID, src, 1, ActivateAbilityParams{}); !errors.Is(err, ErrConditionNotMet) {
		t.Errorf("controller off its turn: err = %v, want ErrConditionNotMet", err)
	}
	if err := g.ActivateCatalogAbility(active.ID, src, 1, ActivateAbilityParams{}); err != nil {
		t.Errorf("active non-controller: %v", err)
	}
}

// TestAnyPlayerRowStillObeysTheObjectsRestrictions: Arrest's "its
// activated abilities can't be activated" restricts the OBJECT, so it
// stops every player (ADR 0106 §1 decision 2); and a permanent that has
// lost its abilities has no any-player row left to offer (CR 613.1f).
func TestAnyPlayerRowStillObeysTheObjectsRestrictions(t *testing.T) {
	for _, tc := range []struct {
		name string
		mod  func(*Characteristic)
		want error
	}{
		{"arrested", func(e *Characteristic) { e.Restrictions |= CantActivate }, ErrCantActivate},
		{"lost its abilities", func(e *Characteristic) { e.AbilitiesRemoved = true }, ErrCardCallerMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			active, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			var resolvedFor uuid.UUID
			stubAnyPlayerCard(t, &resolvedFor)
			src := pushGateCard(g, "Test Hourglass", "Artifact", anyPlayerOracle, other.ID)
			g.WithWriteLock(func() {
				g.RecomputeLayersIfStaleLocked()
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == src {
						eff := g.Battlefield.Cards[i].Effective()
						tc.mod(&eff)
						g.Battlefield.Cards[i].effective = &eff
					}
				}
			})
			if err := g.ActivateCatalogAbility(active.ID, src, 1, ActivateAbilityParams{}); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// TestMayActivate pins the predicate's two arms.
func TestMayActivate(t *testing.T) {
	me, them := uuid.New(), uuid.New()
	mine := Card{Owner: me, Controller: me}
	theirs := Card{Owner: them, Controller: them}
	plain := ActivatedAbilityShape{}
	anyone := ActivatedAbilityShape{AnyPlayer: true}
	cases := []struct {
		name string
		c    Card
		zone ZoneKind
		ab   ActivatedAbilityShape
		want bool
	}{
		{"my permanent", mine, ZoneBattlefield, plain, true},
		{"their permanent", theirs, ZoneBattlefield, plain, false},
		{"their permanent, any-player row", theirs, ZoneBattlefield, anyone, true},
		// CR 108.4a off the battlefield: the owner, whatever the row says.
		{"their card in a graveyard, any-player row", theirs, ZoneGraveyard, anyone, false},
		{"my card in my hand", mine, ZoneHand, plain, true},
	}
	for _, tc := range cases {
		if got := MayActivate(me, tc.c, tc.zone, tc.ab); got != tc.want {
			t.Errorf("%s: MayActivate = %v, want %v", tc.name, got, tc.want)
		}
	}
}
