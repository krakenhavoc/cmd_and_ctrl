package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atraxa's Fall — Sorcery {1}{G} (March of the Machine):
//
//	"Destroy target artifact, battle, enchantment, or creature with
//	 flying."
//
// "With flying" qualifies the creature alone: an artifact, a battle or
// an enchantment is a legal target whether or not it flies. Flying is read
// through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22677153-bd87-4ff8-af92-ced4da2aac6c",
		Name:         "Atraxa's Fall",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, battle, enchantment, or creature with flying",
			func(g *game.Game, caster uuid.UUID, c game.Card) bool {
				return c.IsArtifact() || c.IsBattle() || c.IsEnchantment() ||
					(c.IsCreature() && HasKeyword("flying")(g, caster, c))
			}),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalTarget(ctx.Game, item)
		},
	})
}
