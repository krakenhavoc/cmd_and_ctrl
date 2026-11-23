package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ascend from Avernus — Sorcery {X}{W}{W}{W} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Return all creature and planeswalker cards with mana value X or less from
//	 your graveyard to the battlefield. Exile Ascend from Avernus."
//
// X is the announced X (XMatters). Every creature and planeswalker card of the
// caster's graveyard with mana value X or less (read as it sits in the
// graveyard: a card's {X} is 0 there) comes back as one event under their
// control (CR 603.6a). Then the spell exiles itself as it resolves, Reunion of
// the House's closing step, so it never reaches the graveyard (and a copy has
// no card to exile, CR 707.10).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "451e8ece-7389-4d52-8bc9-450cb5e53e6a",
		Name:         "Ascend from Avernus",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				x := ctx.X()
				ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return (c.IsCreature() || c.IsPlaneswalker()) && !c.IsToken() && c.ManaValue() <= x
				})
				if len(ids) == 0 {
					return nil
				}
				return effects.PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, ctx.Controller(), false)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}
