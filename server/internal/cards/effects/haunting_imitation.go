package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haunting Imitation — Sorcery {2}{U} (Innistrad: Crimson Vow):
//
//	"Each player reveals the top card of their library. For each creature
//	 card revealed this way, create a token that's a copy of that card,
//	 except it's 1/1, it's a Spirit in addition to its other types, and it
//	 has flying. If no creature cards were revealed this way, return
//	 Haunting Imitation to its owner's hand."
//
// Players reveal in turn order. The cards stay on their libraries; the
// tokens copy them where they are and enter together under the caster's
// control. Not one creature card revealed returns the spell to hand
// instead of the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c96ed5b2-6286-4f54-becb-c07f54335e63",
		Name:         "Haunting Imitation",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var creatures []uuid.UUID
			for _, p := range apnapLivingSeats(ctx.Game) {
				for _, id := range ctx.Game.RevealTopOfLibraryForEffect(p, item.SourceCardID, 1, "Haunting Imitation — reveals the top card of their library") {
					if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
						creatures = append(creatures, id)
					}
				}
			}
			if len(creatures) == 0 {
				return ctx.Game.BounceToHandForEffect(item.SourceCardID)
			}
			return CreateTokenCopiesTogether{Copies: creatures, Except: vowB17SpiritCopyException(true)}.Apply(ctx)
		},
	})
}
