package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vizier of Deferment — {2}{W} Creature — Human Cleric 2/2 (Amonkhet, #5535):
//
//	"Flash
//	 When this creature enters, you may exile target creature if it attacked
//	 or blocked this turn. Return that card to the battlefield under its
//	 owner's control at the beginning of the next end step."
//
// Any creature may be targeted; whether it attacked or blocked this turn
// is read as the ability resolves, and a creature that did neither is left
// alone (CR 608.2c). The return is the shared next-end-step flicker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "008d5896-6fc9-4aaa-8c6f-a44c2feb98bb",
		Name:            "Vizier of Deferment",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Vizier of Deferment — exile target creature if it attacked or blocked this turn, return it at the next end step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard && effects.S82OldB099AttackedOrBlockedThisTurn(g, t.ID) {
								return effects.ExileTargetsThenScheduleReturn(ctx, "Vizier of Deferment — return the exiled card")
							}
						}
						return nil
					}),
				effects.TargetCreature("target creature")),
				"Vizier of Deferment — exile target creature if it attacked or blocked this turn?"),
		},
	})
}
