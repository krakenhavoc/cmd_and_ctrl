package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempt with Reflections — Sorcery {3}{U} (Commander 2013, S82 OLD b387):
//
//	"Tempting offer — Choose target creature you control. Create a token that's
//	 a copy of that creature. Each opponent may create a token that's a copy of
//	 that creature. For each opponent who does, create a token that's a copy of
//	 that creature."
//
// Tempt with Vengeance's tempting-offer helper: the caster's own copy first,
// then each opponent in turn order is asked, and one who accepts gets a copy
// under their control and gives the caster one more. Each copy is of the chosen
// creature as it was copiable (CR 707.2) and is created by the player who
// makes it. A target that has left the battlefield by resolution makes the
// spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	tempt := effects.NewCmdr2S49Tempting("confirm/tempt-with-reflections",
		"Tempt with Reflections — create a token that's a copy of the creature (the caster gets one more if you do)?", temptWithReflectionsCopy)
	effects.Register(effects.Spec{
		OracleID:     "76c142ef-0f07-4215-8d11-d25f7114c70d",
		Name:         "Tempt with Reflections",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return tempt.Resolve(ctx)
		},
	})
}

func temptWithReflectionsCopy(ctx *effects.Context, player uuid.UUID) error {
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.CreateTokenCopy{Controller: player, Copy: id, N: 1}.Apply(ctx)
}
