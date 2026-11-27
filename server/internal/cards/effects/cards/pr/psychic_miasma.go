package cardspr

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Miasma — Sorcery {1}{B} (Scars of Mirrodin, S82 OLD batch 471,
// #5577):
//
//	"Target player discards a card. If a land card is discarded this way,
//	 return Psychic Miasma to its owner's hand."
//
// Mind Roots' discard with a continuation: the target chooses the card, and
// the clause reads what really left their hand. The spell goes to its owner's
// hand as its own resolution ends instead of to the graveyard (Unexpected
// Results' exit), so it is not a return from the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43d69e13-27df-43a9-b610-6628dc02e3b4",
		Name:         "Psychic Miasma",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        1,
				Question: "Psychic Miasma — discard a card",
				Asked:    game.NewAskedItem(item),
			}, psychicMiasmaLand, nil)
		},
	})
}

// psychicMiasmaLand is "if a land card is discarded this way, return Psychic
// Miasma to its owner's hand". An on-disk key: never renamed, never reused.
var psychicMiasmaLand = game.RegisterDiscardThen("choose-cards/s82-old-b471-psychic-miasma-land", func(g *game.Game, r game.DiscardsDone) error {
	effects.OldB471ReturnThisSpellIfLandAmong(g, r.Source, r.Discarded.Cards())
	return nil
})
