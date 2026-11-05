package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Erica Ortegas, Plucky Pilot — {3}{R} Legendary Creature — Human Pilot,
// 3/3 (Star Trek, #4973):
//
//	"Haste
//	 Whenever Erica Ortegas becomes tapped, you may put a land card from
//	 your hand onto the battlefield. If you do, return a land you control
//	 to its owner's hand at the beginning of the next end step."
//
// "Becomes tapped" is ThisBecameTapped over both event kinds (the engine
// taps an attacker without an EventTapCard; a vigilance attacker did not
// become tapped). The put is Spelunking's "you may put a land card from
// your hand" (MayPutALandFromHand); its keyed tail is told what entered,
// and only a land that really arrived schedules the CR 603.7 delayed
// trigger. That trigger names no land: as it resolves its controller
// chooses any land they control then and returns it (ReturnOneYouControl,
// the karoos' resolution-time choice), the land that was put or another.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b2514e5-4124-4f6c-b6b7-24c0ab3d3971",
		Name:            "Erica Ortegas, Plucky Pilot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Erica Ortegas — you may put a land card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					put := MayPutALandFromHand("Erica Ortegas")
					put.ThenTail = ericaOrtegasPut
					put.ThenItem = true
					return put.Apply(NewContext(g, item))
				}),
		},
	})
}

// ericaOrtegasPut is "If you do, return a land you control to its owner's
// hand at the beginning of the next end step", keyed. An on-disk key.
var ericaOrtegasPut = PutFromHandTail("erica-ortegas/return-a-land-at-end", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:      "Erica Ortegas — return a land you control to its owner's hand",
		Controller: res.Player,
		Body:       ericaOrtegasReturnLandBody,
	}.Apply(ctx)
})

// ericaOrtegasReturnLandBody is the delayed trigger: its controller chooses
// a land they control as it resolves and returns it to its owner's hand.
var ericaOrtegasReturnLandBody = game.SimpleDelayedBody("erica-ortegas/return-a-land-you-control", func(g *game.Game, item *game.StackItem) error {
	return ReturnOneYouControl{
		Match:    MatchLand,
		Question: "Erica Ortegas — return a land you control to its owner's hand",
	}.Apply(NewContext(g, item))
})
