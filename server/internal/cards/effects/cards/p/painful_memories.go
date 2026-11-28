package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Painful Memories — Sorcery {1}{B} (Mirage, S82 OLD b798, #5633):
//
//	"Look at target opponent's hand and choose a card from it. Put that card
//	 on top of that player's library."
//
// Vendilion Clique's look-at-a-hand pick (only the caster sees the hand,
// CR 701.20e; the chosen card stays in it for the continuation), mandatory
// and with any card choosable. The continuation puts the chosen card on top of
// its owner's library. An empty hand does nothing (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27a3faa1-4f1a-4581-9e19-9265faa9c10b",
		Name:         "Painful Memories",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Label:  "card",
				Look:   true,
				Keep:   true,
				Then:   painfulMemoriesTuck,
			}.Apply(ctx)
		},
	})
}

// painfulMemoriesTuck puts the chosen card on top of its owner's library. An
// on-disk key: never renamed, never reused.
var painfulMemoriesTuck = effects.RevealedPickThen("revealed-pick/s82-old-b798-painful-memories-top", func(ctx *effects.Context, pick game.RevealedPick) error {
	if len(pick.Chosen) == 0 {
		return nil
	}
	return effects.PutIntoLibrary{Card: pick.Chosen[0].InstanceID}.Apply(ctx)
})
