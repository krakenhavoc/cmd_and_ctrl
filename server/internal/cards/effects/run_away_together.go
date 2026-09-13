package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Run Away Together — Instant {1}{U} (EDHREC rank 1680):
//
//	"Choose two target creatures controlled by different players.
//	 Return those creatures to their owners' hands."
//
// Two mana to save your own creature and bounce someone else's — or
// to reset two opponents' best threats at once. Two target slots
// over creatures; each is re-checked at resolution, and a creature
// that left in response is skipped while the other still returns
// (CR 608.2b).
//
// Sandbox simplification: "controlled by different players" is a
// constraint ACROSS the two slots, and a target clause is checked
// one slot at a time — there is no way to refuse the second pick
// because of the first. So the pair is checked at resolution
// instead: two creatures still legal and controlled by the same
// player is not a legal pair, and the spell does nothing. Weaker
// than printed (a mis-picked cast is wasted rather than refused),
// never stronger — no line exists that the printed card forbids.
func init() {
	Register(Spec{
		OracleID:     "290faa28-450e-4797-9a8f-642d8af3f82a",
		Name:         "Run Away Together",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The two creatures must have different controllers, but that isn't checked until the spell resolves — a pair with the same controller is accepted and the spell then does nothing."},
		Targets:      TargetCreature("two target creatures controlled by different players").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			controllers := map[uuid.UUID]bool{}
			for _, t := range legal {
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					controllers[c.Controller] = true
				}
			}
			if len(legal) == 2 && len(controllers) < 2 {
				return nil
			}
			for _, t := range legal {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
