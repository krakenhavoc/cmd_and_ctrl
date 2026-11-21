package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Interceptor Mechan — Artifact Creature — Robot {2}{B}{R}, 2/2 (Edge of
// Eternities):
//
//	"Flying
//	 When this creature enters, return target artifact or creature card
//	 from your graveyard to your hand.
//	 Void — At the beginning of your end step, if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn, put
//	 a +1/+1 counter on this creature."
//
// The enters trigger is mandatory and targets as it goes on the stack, so
// it is not put there at all with no artifact or creature card in your
// graveyard (CR 603.3d); the target is re-checked as it resolves (CR
// 608.2b). Void is an intervening if (void.go). Flying is printed.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Interceptor Mechan — return target artifact or creature card from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, id := range effects.LegalTargetIDs(ctx) {
				return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
			}
			return nil
		})
	enters.Targets = effects.TargetCardInGraveyard("target artifact or creature card in your graveyard",
		effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature()))
	effects.Register(effects.Spec{
		OracleID:        "ded2b1b7-8c12-4f6f-a3be-a8f7b62b0819",
		Name:            "Interceptor Mechan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			enters,
			effects.AtYourEndStepIfVoid("Interceptor Mechan — put a +1/+1 counter on this creature", effects.PutACounterOnThis("+1/+1")),
		},
	})
}
