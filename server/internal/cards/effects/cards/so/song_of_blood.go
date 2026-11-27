package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Song of Blood — Sorcery {1}{R} (Visions, S82 older sets batch 789, #5631):
//
//	"Mill four cards. Whenever a creature attacks this turn, it gets +1/+0
//	 until end of turn for each creature card put into your graveyard this
//	 way."
//
// The mill's keyed tail counts the creature cards that actually landed in the
// graveyard (a card a replacement sent elsewhere is not counted), then sets up
// a repeating delayed trigger (CR 603.7b): each creature that attacks this
// turn, whoever controls it, gets that many +1/+0 as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab4b6524-3667-4db3-b5cb-1b02ee944906",
		Name:         "Song of Blood",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(effects.Old789SongOfBloodMilled)}.Apply(ctx)
		},
	})
}
