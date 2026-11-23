package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Purge — Sorcery {2}{W} (Innistrad Remastered):
//
//	"As an additional cost to cast this spell, sacrifice a permanent.
//	 Exile target artifact, creature, or enchantment."
//
// The sacrifice is paid as the spell is cast, so the permanent can't also be
// the target (the target is chosen first, but a sacrificed target is gone and
// the spell does nothing, CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b262a057-9ace-48b4-b418-1bb751ccb300",
		Name:           "Angelic Purge",
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a permanent"),
		Targets:        effects.TargetPermanent("target artifact, creature, or enchantment", effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
