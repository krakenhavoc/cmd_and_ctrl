package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent Maria Hill — Legendary Creature — Human Spy Hero {W}, 2/1
// (Marvel Super Heroes, #4455, seam #4534):
//
//	"Whenever Agent Maria Hill becomes tapped to pay a teamwork cost, put
//	 a +1/+1 counter on her and draw a card."
//
// The teamwork payment stamps its EventTapCard with Event.TapCause
// (game.TapCauseTeamwork), and ThisTappedToPayTeamwork reads it: an
// ordinary tap, an attack or a crew tap does nothing. The counter and the
// draw both happen as the ability resolves; a Maria Hill that has left the
// battlefield or come back as a new object (CR 400.7) gets no counter, and
// her controller still draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2d90efe-80f2-4069-b135-a8bd00ccd2b0",
		Name:         "Agent Maria Hill",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventTapCard, ThisTappedToPayTeamwork,
				"Agent Maria Hill — put a +1/+1 counter on her and draw a card",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, DrawCards{N: 1})),
		},
	})
}
