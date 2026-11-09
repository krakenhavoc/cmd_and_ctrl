package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kav Landseeker — Creature — Kavu Soldier {3}{R}, 4/3 (Edge of
// Eternities, #4461):
//
//	"Menace
//	 When this creature enters, create a Lander token. At the beginning
//	 of the end step on your next turn, sacrifice that token."
//
// Harried Dronesmith's keyed creation: the sacrifice is the creation's
// continuation, told the tokens that were really made (a doubler makes
// two, and both are "that token", per the rulings). A delayed trigger
// (CR 603.7) sacrifices exactly those tokens at the beginning of the end
// step of the controller's NEXT turn: AfterThisTurn is the seam that makes
// "next" strict, so a Landseeker cast in its controller's own turn does
// not sacrifice the Lander at that turn's end step. A Lander someone else
// has taken or that is already gone is not sacrificed, and sacrificing it
// does not run its search. Menace is a printed keyword.
//
// Seam: ScheduleDelayedTrigger.AfterThisTurn (DelayedTrigger.NotBeforeSeq).
func init() {
	Register(Spec{
		OracleID:        "528fe479-dab9-4cf8-bda0-1d226a644366",
		Name:            "Kav Landseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kav Landseeker — create a Lander token; sacrifice it at the end step on your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(b16LanderToken()), Count: 1}},
					}, ctx.ItemTail(kavLandseekerSacrificeLater))
				}),
		},
	})
}

// kavLandseekerSacrificeLater schedules the sacrifice of the tokens the
// creation made. An on-disk key.
var kavLandseekerSacrificeLater = CardTail("kav-landseeker/sacrifice-the-lander-next-turn", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:              "Kav Landseeker — sacrifice the Lander",
		ControllerTurnOnly: true,
		AfterThisTurn:      true,
		Cards:              o.IDs,
		Body:               sacrificeListedCardsBody,
	}.Apply(ctx)
})
