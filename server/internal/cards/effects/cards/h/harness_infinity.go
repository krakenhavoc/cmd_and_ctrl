package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harness Infinity — Instant {1}{B}{B}{B}{G}{G}{G} (Strixhaven: School of
// Mages, S82 batch 047, #5502):
//
//	"Exchange your hand and graveyard.
//	 Exile Harness Infinity."
//
// The exchange (CR 701.12) is a swap of two zones' contents decided up
// front: the cards in the graveyard and the cards in the hand are both
// read before either moves, then the graveyard's go to the hand and the
// hand's go to the graveyard. Nothing is discarded (the hand's cards are
// put, not discarded) and nothing is returned "from" anywhere a trigger
// would read as a discard. A commander either way is offered the command
// zone (CR 903.9) by the move. Harness Infinity is still on the stack while
// this happens, so it is in neither zone; it then moves itself to exile
// (Avatar's Wrath's last line), so a countered one stays in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efd31076-2bbe-4c65-9166-19ec0e832b57",
		Name:         "Harness Infinity",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := item.Controller
			p := ctx.Game.PlayerByIDForEffect(player)
			if p == nil {
				return nil
			}
			hand := effects.AllHandCardIDs(ctx.Game, player)
			graveyard := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
			for _, c := range p.Graveyard.Cards {
				graveyard = append(graveyard, c.InstanceID)
			}
			for _, id := range graveyard {
				if err := ctx.Game.ReturnFromGraveyardForEffect(id, game.ZoneHand); err != nil {
					return err
				}
			}
			for _, id := range hand {
				if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
					return err
				}
			}
			// "Exile Harness Infinity."
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
