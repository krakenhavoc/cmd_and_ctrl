package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Council's Deliberation — Instant {1}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Draw a card.
//	 Whenever you scry, if you control an Island, you may exile this card from
//	 your graveyard. If you do, draw a card."
//
// The spell is a plain draw. The second ability works from the graveyard
// (CR 113.6k, the InGraveyard wrapper: "you" is the card's owner while it
// sits there). Its intervening "if" (CR 603.4) is checked as the scry happens
// and again as the trigger resolves, reading the Islands you control through
// the layers. It is a "may" asked as the trigger resolves; on a yes the card
// is exiled from the graveyard and only if it really was exiled, you draw
// ("if you do"). A card that has left the graveyard in response, or come
// back as a new object, is not exiled and draws nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3280488d-e94b-4d53-b279-c32ed711d4bc",
		Name:         "Council's Deliberation",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.DrawCards{N: 1}),
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.Optional(
				effects.InterveningIf(effects.CondBoard{
					Name: "you control an Island",
					Test: func(ctx *effects.Context) bool {
						return effects.ControlsA(ctx.TextSubtype("Island"))(ctx.Game, ctx.Controller())
					},
				}, effects.On(game.EventScry, effects.ByYou,
					"Council's Deliberation — exile this card from your graveyard to draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if effects.SourceIsNewObject(g, item) {
							return nil
						}
						z := g.FindCardZoneForEffect(item.SourceCardID)
						if z == nil || z.Kind != game.ZoneGraveyard {
							return nil
						}
						return effects.ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(councilsDeliberationDraw)}.Apply(ctx)
					})),
				"Council's Deliberation — exile this card from your graveyard to draw a card?")),
		},
	})
}

// councilsDeliberationDraw is "if you do, draw a card": it runs once the card
// has reached exile. An on-disk key: never renamed, never reused.
var councilsDeliberationDraw = effects.CardTail("then/councils-deliberation-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
