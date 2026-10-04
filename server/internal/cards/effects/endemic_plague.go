package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endemic Plague — Sorcery {3}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Destroy all creatures that share a creature type with the
//	 sacrificed creature. They can't be regenerated."
//
// The creature types are the sacrificed creature's as it last existed
// on the battlefield (CR 608.2h), read off the payment record (ADR 0113
// §1). A changeling on either side shares a type with every creature
// that has one (CR 702.73a), and a creature with no creature type shares
// nothing. The destruction ignores regeneration shields (CR 701.19c).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "db982577-1c75-4bc9-ab15-1888ea0be16d",
		Name:           "Endemic Plague",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			info, ok := ctx.SacrificedPermanent()
			if !ok {
				return nil
			}
			return DestroyAllMatching{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && sharesACreatureTypeWith(info, c)
				},
				CantBeRegenerated: true,
			}.Apply(ctx)
		},
	})
}
