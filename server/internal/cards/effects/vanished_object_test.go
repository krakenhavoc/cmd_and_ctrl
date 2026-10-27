package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// testNothingTail is a keyed "then" that does nothing, for a door whose
// tail is beside the point (#4253 retired the closure Then these cases
// used to pass).
var testNothingTail = CardTail("test/4253-nothing", func(*Context, game.TailOutcome) error { return nil })

// #2620: an ability's instruction that names a single object — its own
// source ("put a counter on this", "tap this") or its one target — when
// that object has ceased to exist (a token that died, a card that
// vanished in response) does nothing (CR 608.2b). It is not an effect
// error: the soak counts EventEffectError as a failure and a real table
// would log noise for a legal line of play. Every primitive that takes
// a single object belongs in this table.
func TestPrimitivesOnAVanishedObjectDoNothing(t *testing.T) {
	cases := []struct {
		name string
		prim func(gone uuid.UUID) Applier
	}{
		{"AddCounter", func(id uuid.UUID) Applier { return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1} }},
		{"AddCounter remove", func(id uuid.UUID) Applier { return AddCounter{Target: id, Kind: game.CounterPlusOne, N: -1} }},
		{"TapTarget", func(id uuid.UUID) Applier { return TapTarget{Target: id} }},
		{"UntapTarget", func(id uuid.UUID) Applier { return UntapTarget{Target: id} }},
		{"DestroyTarget", func(id uuid.UUID) Applier { return DestroyTarget{Target: id} }},
		{"Regenerate", func(id uuid.UUID) Applier { return Regenerate{Target: id} }},
		{"SacrificePermanent", func(id uuid.UUID) Applier { return SacrificePermanent{Target: id} }},
		{"SacrificePermanent ThenTail", func(id uuid.UUID) Applier {
			return SacrificePermanent{Target: id, ThenTail: game.KeyedTail{Ref: testNothingTail}}
		}},
		{"ExileTarget", func(id uuid.UUID) Applier { return ExileTarget{Target: id} }},
		{"Flicker", func(id uuid.UUID) Applier { return Flicker{Target: id} }},
		{"ReturnFromExile", func(id uuid.UUID) Applier { return ReturnFromExile{Target: id} }},
		{"BounceToHand", func(id uuid.UUID) Applier { return BounceToHand{Target: id} }},
		{"BounceToHand ThenTail", func(id uuid.UUID) Applier {
			return BounceToHand{Target: id, ThenTail: game.KeyedTail{Ref: testNothingTail}}
		}},
		{"ReturnFromGraveyard", func(id uuid.UUID) Applier { return ReturnFromGraveyard{Target: id} }},
		{"PhaseOut", func(id uuid.UUID) Applier { return PhaseOut{Targets: []uuid.UUID{id}} }},
		{"CounterTarget", func(id uuid.UUID) Applier { return CounterTarget{StackID: id} }},
		{"ReturnSpellToHand", func(id uuid.UUID) Applier { return ReturnSpellToHand{StackID: id} }},
		{"ExileTargetSpell", func(id uuid.UUID) Applier { return ExileTargetSpell{StackID: id} }},
		{"ExileTargetSpell ThenTail", func(id uuid.UUID) Applier {
			return ExileTargetSpell{StackID: id, ThenTail: game.KeyedTail{Ref: testNothingTail}}
		}},
		{"BecomePrepared", func(id uuid.UUID) Applier { return BecomePrepared{Target: id} }},
		{"DealDamage", func(id uuid.UUID) Applier { return DealDamage{Source: id, Target: id, Amount: 2} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			gone := uuid.New()
			item := &game.StackItem{Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: gone}
			var err error
			g.WithWriteLock(func() { err = tc.prim(gone).Apply(NewContext(g, item)) })
			if err != nil {
				t.Errorf("%s on a vanished object: %v, want nil", tc.name, err)
			}
		})
	}
}
