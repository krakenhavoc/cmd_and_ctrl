package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Architects of Will — {2}{U}{B} Artifact Creature — Human Wizard 3/3 (Archenemy):
//
//	"When this creature enters, look at the top three cards of target player's
//	 library, then put them back in any order.
//	 Cycling {U/B}"
func init() {
	trigger := effects.WhenThisEnters("Architects of Will — look at the top three cards of target player's library, then put them back in any order",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return effects.LookAtLibraryThenPlace{
					Owner:     t.ID,
					N:         3,
					Placement: game.LibraryPlaceTop,
					Label:     "Architects of Will — put them back in any order",
				}.Apply(ctx)
			}
			return nil
		})
	trigger.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "79d290ea-c095-4135-9711-f3ce089dae43",
		Name:         "Architects of Will",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{U/B}")},
	})
}
