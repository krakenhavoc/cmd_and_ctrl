package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Funeral — Sorcery {1}{U}{B} (Alara Reborn, #5586):
//
//	"Target opponent reveals cards from the top of their library until four
//	 land cards are revealed. That player puts all cards revealed this way into
//	 their graveyard."
//
// Mirko Vosk's reveal-and-mill with a stop at the fourth land (the run is
// measured and announced as one reveal, then milled as one Until run, the fourth
// land milled too). A library with fewer than four lands reveals and mills
// everything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf5103c1-2590-4aed-b3ae-8a570f797738",
		Name:         "Mind Funeral",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return effects.S82OldB405MillUntilFourLands(ctx, victim)
		},
	})
}
