package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prophecy — Sorcery {W} (Homelands, S82 OLD b826, #5635):
//
//	"Reveal the top card of target opponent's library. If it's a land, you
//	 gain 1 life. Then that player shuffles.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The top card is revealed (not drawn, CR 701.20a), a land gives the caster 1
// life, the opponent shuffles their library whether or not a card was there,
// and the draw is the delayed trigger for the next upkeep at the table
// (CR 603.7a, next_upkeep_draw.go). With its one target gone the spell does not
// resolve, so no draw is set up (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8dcf15fd-f000-4424-b74c-22fc2642b8c3",
		Name:         "Prophecy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp := effects.FirstLegalTargetPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			var revealed []uuid.UUID
			if err := (effects.RevealTopOfLibrary{
				Player:   opp,
				N:        1,
				Reason:   "Prophecy — reveal the top card of target opponent's library",
				Revealed: &revealed,
			}).Apply(ctx); err != nil {
				return err
			}
			if len(revealed) > 0 {
				if c, ok := ctx.Game.LookupCardForEffect(revealed[0]); ok && c.HasCardType("land") {
					if err := (effects.GainLife{Player: ctx.Controller(), Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if err := ctx.Game.ShuffleLibraryForEffect(opp); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Prophecy — draw a card").Apply(ctx)
		},
	})
}
