package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	enters := WhenThisEnters("Interceptor Mechan — return target artifact or creature card from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, id := range legalTargetIDs(ctx) {
				return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
			}
			return nil
		})
	enters.Targets = TargetCardInGraveyard("target artifact or creature card in your graveyard",
		YouOwn(), Or(Artifact(), Creature()))
	Register(Spec{
		OracleID:        "ded2b1b7-8c12-4f6f-a3be-a8f7b62b0819",
		Name:            "Interceptor Mechan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			enters,
			AtYourEndStepIfVoid("Interceptor Mechan — put a +1/+1 counter on this creature", putACounterOnThis("+1/+1")),
		},
	})
}
