package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stall Out — Sorcery {1}{U} (Aetherdrift, #4464):
//
//	"Tap target creature or Vehicle, then put three stun counters on it.
//	 (If a permanent with a stun counter would become untapped, remove one
//	 from it instead.)
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Cryotheory Adept's tap-and-stun with three counters, the target a
// creature or a Vehicle, crewed or not (Bounce Off's predicate). The
// counters are the engine's stun replacement (CR 122.1d): each untap the
// permanent would get removes one instead. A target gone by resolution
// gets nothing (CR 608.2b). Cycling is the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce55c0a9-3334-440e-ae1b-c04a9e1c3d87",
		Name:         "Stall Out",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (TapTarget{Target: t}).Apply(ctx); err != nil {
				return err
			}
			return AddCounter{Target: t, Kind: game.CounterStun, N: 3}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
