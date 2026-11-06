package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clone Legion — Sorcery {7}{U}{U} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"For each creature target player controls, create a token that's a
//	 copy of that creature."
//
// Saheeli, the Gifted's −7 shape (CreateTokenCopiesTogether): the
// creatures the target player controls as the spell resolves, one token
// copying each, all created at once (CR 608.2f) and entering together, so
// each sees the others enter and every copied enters-ability triggers (the
// rulings). Each token copies the copiable values only (CR 707.2): not
// tapped status, counters, Auras or other effects; a copied token or a
// creature copying something copies what that one copies, a face-down
// creature makes a face-up nameless 2/2 (CR 707.2, 708.2), and {X} is zero
// (CR 707.9a). The legend rule applies afterwards as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "690bd111-5b32-4156-b650-9afad3d6ef3d",
		Name:         "Clone Legion",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == player && c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return CreateTokenCopiesTogether{Copies: ids}.Apply(ctx)
		},
	})
}
