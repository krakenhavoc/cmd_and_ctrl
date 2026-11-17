package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farsight Mask — Artifact {5} (Mirrodin, #5606):
//
//	"Whenever a source an opponent controls deals damage to you, if this
//	 artifact is untapped, you may draw a card."
//
// Michiko Konda's damage watch: one trigger per damage event to the
// Mask's controller (CR 603.2), so three attackers connecting in one
// combat damage step are three draws, and damage to their permanents
// does not count. "If this artifact is untapped" is an intervening if
// (CR 603.4): checked as the ability triggers and again as it resolves,
// before the "you may" is asked, so tapping the Mask in response stops
// the draw with no question.
//
// No simplification.
func init() {
	draw := OptionalDraw(On(game.EventDealDamage, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return !source.Tapped && ASourceAnOpponentControlsDealtDamageToYou(ev, source, lki, g)
	}, "Farsight Mask — you may draw a card", Do(DrawCards{N: 1})), "Draw a card with Farsight Mask?", 1)
	Register(Spec{
		OracleID:     "c82091b2-f2f1-4a85-80d6-91ab218f0a94",
		Name:         "Farsight Mask",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{Name: "this artifact is untapped", Test: func(ctx *Context) bool {
				return b26SourceOnBattlefieldUntapped(ctx.Game, ctx.Item)
			}}, draw),
		},
	})
}
