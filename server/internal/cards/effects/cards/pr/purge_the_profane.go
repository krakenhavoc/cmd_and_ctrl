package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Purge the Profane — Sorcery {2}{W}{B} (Gatecrash, S82 OLD b408, #5565):
//
//	"Target opponent discards two cards and you gain 2 life."
//
// Gerrard's Verdict's discard (the opponent chooses, CR 701.9a) with the life
// gained as the discard's continuation, so it waits for the discard to finish,
// and is gained even when the opponent had no cards to discard (the sentence's
// two halves are independent). A target that is gone fizzles the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0f29566-bca8-4fe8-86ce-4a051ece6d5a",
		Name:         "Purge the Profane",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Purge the Profane — discard two cards",
				Asked:    game.NewAskedItem(item),
			}, purgeTheProfaneLife, nil)
		},
	})
}

// purgeTheProfaneLife is "you gain 2 life" once the discard is done. An
// on-disk key: never renamed, never reused.
var purgeTheProfaneLife = game.RegisterDiscardThen("choose-cards/s82-old-b408-purge-the-profane-life", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	return effects.GainLife{Player: r.Item.Controller, Amount: 2}.Apply(effects.NewContext(g, r.Item))
})
