package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immortal Servitude — Sorcery {X}{W/B}{W/B}{W/B} (Gatecrash, #5565):
//
//	"Return each creature card with mana value X from your graveyard to the
//	 battlefield."
//
// Rally the Ancestors' graveyard sweep with an equality: the set is chosen as
// the spell resolves and returns together (ReturnFromGraveyardTogether). The
// cards' mana values are read in the graveyard, where an {X} in a cost is 0
// (CR 107.3), so X = 0 returns the creature cards whose cost has no generic
// or coloured mana beyond X, and tokens never qualify.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c946be2c-1e8f-46bc-ab75-861926d4aa63",
		Name:         "Immortal Servitude",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
				return c.IsCreature() && !c.IsToken() && c.ManaValue() == x
			})
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
		},
	})
}
