package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage's Knowledge — Sorcery {2}{U} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Return target sorcery card from your graveyard to your hand."
//
// Raise Dead's return for a sorcery card in the caster's own graveyard,
// checked again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8ffd53b-5b10-4866-83fe-630eee58c514",
		Name:         "Sage's Knowledge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target sorcery card from your graveyard", effects.YouOwn(), effects.Sorcery()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
