package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inferno Hellion — Creature — Hellion {3}{R}, 7/3 (Jumpstart, #5510):
//
//	"Trample
//	 At the beginning of each end step, if this creature attacked or blocked
//	 this turn, its owner shuffles it into their library."
//
// An intervening "if" (CR 603.4), checked as the end step begins and again
// as the ability resolves: the Hellion attacked this turn (the per-object
// attack tally) or was declared as a blocker since it last entered (this
// turn's block events; a Hellion that left and came back is a new object,
// CR 400.7). It is shuffled into its owner's library (not its controller's)
// and that owner shuffles; the ability does nothing for a Hellion that has
// already left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68c88a4c-8c2b-4e6f-b835-7e269eb073c8",
		Name:            "Inferno Hellion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "this creature attacked or blocked this turn",
				Test: func(ctx *effects.Context) bool {
					return effects.S82OldB099AttackedOrBlockedThisTurn(ctx.Game, ctx.Source())
				},
			}, effects.On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB099AttackedOrBlockedThisTurn(g, source.InstanceID)
			}, "Inferno Hellion — its owner shuffles it into their library", effects.S82OldB099ShuffleThisIntoItsOwnersLibrary)),
		},
	})
}
