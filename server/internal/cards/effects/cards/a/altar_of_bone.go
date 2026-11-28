package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar of Bone — Sorcery {G}{W} (Ice Age, S82 OLD, #5636):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Search your library for a creature card, reveal it, put it into your hand,
//	 then shuffle."
//
// Ritual of the Machine's sacrifice cost with Fauna Shaman's search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "607a3f5f-4a36-4077-b5cf-40ffd52fa580",
		Name:           "Altar of Bone",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerSacOutlet},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.MatchCreature,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Altar of Bone — a creature card, to your hand",
			}.Apply(ctx)
		},
	})
}
