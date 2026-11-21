package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Order of Whiteclay — Creature — Kithkin Cleric {1}{W}{W}, 1/4 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"{1}{W}{W}, {Q}: Return target creature card with mana value 3 or less
//	 from your graveyard to the battlefield. ({Q} is the untap symbol.)"
//
// {Q} is AbilityCost.UntapSelf (CR 107.6): the Order must be tapped to pay
// it, and a creature that has not been under its controller's control since
// their turn began cannot (CR 302.6). The reanimation is Balduvian
// Atrocity's graveyard return; a target gone by resolution makes the ability
// do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16c46eaf-7a31-4957-ab7c-ff07fbbd9ddc",
		Name:         "Order of Whiteclay",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{W}, {Q}: Return target creature card with mana value 3 or less from your graveyard to the battlefield.",
			Cost:    game.AbilityCost{Mana: "{1}{W}{W}", UntapSelf: true},
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Targets: effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
				effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ReturnFromGraveyardTogether{
					Targets:    effects.LegalTargetCardIDs(ctx),
					Controller: ctx.Controller(),
				}.Apply(ctx)
			},
		}},
	})
}
