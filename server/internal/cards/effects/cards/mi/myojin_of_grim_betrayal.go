package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myojin of Grim Betrayal — Legendary Creature — Spirit {5}{B}{B}{B},
// 5/2 (Kamigawa: Neon Dynasty Commander, #5493):
//
//	"Myojin of Grim Betrayal enters with an indestructible counter on it
//	 if you cast it from your hand.
//	 Remove an indestructible counter from Myojin of Grim Betrayal: Put
//	 onto the battlefield under your control all creature cards in all
//	 graveyards that were put there from anywhere this turn."
//
// The counter is Myojin of Night's Reach's entry clause. The cards are
// found as the ability resolves from this turn's zone moves into a
// graveyard (neoCreatureCardsPutIntoGraveyardsThisTurn: a token is no
// card) and enter together under your control, as Rise of the Dark
// Realms' do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "896fc91a-1f57-45e2-b135-37bc9c2c80a7",
		Name:                       "Myojin of Grim Betrayal",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: effects.NeoMyojinEntry(),
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove an indestructible counter from Myojin of Grim Betrayal: Put onto the battlefield under your control all creature cards in all graveyards that were put there from anywhere this turn.",
			Cost:    effects.NeoRemoveIndestructible(),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ReturnFromGraveyardTogether{
					Targets:    effects.NeoCreatureCardsPutIntoGraveyardsThisTurn(ctx),
					Controller: ctx.Controller(),
				}.Apply(ctx)
			},
		}},
	})
}
