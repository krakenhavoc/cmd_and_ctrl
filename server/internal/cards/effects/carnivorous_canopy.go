package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carnivorous Canopy — Sorcery {2}{G} (Phyrexia: All Will Be One, #5479):
//
//	"Destroy target artifact, enchantment, or creature with flying. If that
//	 permanent's mana value was 3 or less, proliferate."
//
// The mana value is read before the destroy (CR 608.2h), and the
// proliferate happens even if the permanent survives (indestructible).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3360259-19df-43c6-95a9-e76770f2fb77",
		Name:         "Carnivorous Canopy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			target := item.Targets[0].ID
			// The proliferate is the destroy's keyed tail: it runs once the
			// destruction has settled (a replacement may pause it) and whether
			// or not the permanent was destroyed (indestructible).
			var then game.KeyedTail
			if s82OneB08ManaValueAtMost(ctx, target, 3) {
				then = ctx.TailN(thenProliferateAgain, 1)
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{target}, then)
		},
	})
}
