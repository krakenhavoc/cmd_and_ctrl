package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlwind Technique — Instant — Lesson {4}{U}{U} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Target player draws two cards, then discards a card.
//	 Airbend up to two target creatures. (To airbend a creature, exile
//	 it. While it's exiled, its owner may cast it for {2} rather than
//	 its mana cost.)"
//
// Two target clauses: slot 0 the player, slot 1 up to two creatures. Each
// clause is checked on its own as the spell resolves (CR 608.2b): a
// player who left does not draw, and an illegal creature is not
// airbended, while the rest still happens. The creatures are airbended
// together, one exile (AirbendAll), each owner getting the {2} cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4477852-66eb-4a54-b4c1-760cbf624f82",
		Name:         "Whirlwind Technique",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("up to two target creatures").WithCount(0, 2),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if p, ok := ctx.ClauseTarget(0); ok && p.Kind == game.TargetPlayer {
				if err := drawThenDiscard(ctx.Game, p.ID, item.SourceCardID, 2, 1, "Whirlwind Technique — discard a card"); err != nil {
					return err
				}
			}
			var creatures []uuid.UUID
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					creatures = append(creatures, t.ID)
				}
			}
			if len(creatures) == 0 {
				return nil
			}
			return AirbendAll{Targets: creatures}.Apply(ctx)
		},
	})
}
