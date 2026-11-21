package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Break Asunder — Sorcery {2}{G}{G} (Scourge):
//
//	"Destroy target artifact or enchantment.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Naturalize at sorcery speed, and cycling as the keyword's own hand
// ability (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2c53737-c265-46e3-a779-52c6b4f82d7d",
		Name:         "Break Asunder",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
