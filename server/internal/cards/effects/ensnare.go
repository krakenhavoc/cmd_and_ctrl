package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ensnare — Instant {3}{U} (Nemesis, #5837):
//
//	"You may return two Islands you control to their owner's hand rather
//	 than pay this spell's mana cost.
//	 Tap all creatures."
//
// Gush's alternative cost (ReturnInstead over a clause of exactly two,
// #5837): two Islands the caster controls go to their owners' hands as a
// cost, with Ensnare already on the stack (CR 118.9, 601.2h, 400.3).
// "Islands" is the land type (CR 205.3i). "All creatures" is every
// creature on the battlefield, whoever controls it, read as Ensnare
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df197279-ddb2-4237-8bb1-8df41d3bed9e",
		Name:         "Ensnare",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			ReturnInstead(
				"Return two Islands you control to their owner's hand",
				PermanentYouControl("two Islands you control", HasSubtype("Island")).WithCount(2, 2),
				"two Islands you control",
			),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, Creature()) {
				ids = append(ids, c.InstanceID)
			}
			return tapEach(ctx, ids)
		},
	})
}
