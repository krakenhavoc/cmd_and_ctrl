package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Submersion — Sorcery {4}{R} (Shards of Alara):
//
//	"Destroy target artifact or land.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A targeted destroy over artifacts and lands, re-judged as it resolves
// (CR 608.2b), and the shared cycling ability from the hand (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37985fef-696e-4c6b-9a60-29c5051b5a25",
		Name:         "Volcanic Submersion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalTarget(ctx.Game, ctx.Item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
