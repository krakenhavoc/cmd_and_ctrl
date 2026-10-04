package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forge Armor — Instant {4}{R}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Put X +1/+1 counters on target creature, where X is the sacrificed
//	 artifact's mana value."
//
// X is the sacrificed artifact's mana value as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1): a
// Treasure or any other token that is no copy is 0, and puts nothing on.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "7d4d9bfa-7e85-461b-bd62-1de97690a110",
		Name:           "Forge Armor",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact", Artifact()),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := sacrificedManaValue(ctx)
			legal := ctx.LegalTargets()
			if x <= 0 || len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			return AddCounter{Target: legal[0].ID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
		},
	})
}
