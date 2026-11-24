package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flunk — Instant {1}{B} (Strixhaven: School of Mages, S82 batch 047, #5502):
//
//	"Target creature gets -X/-X until end of turn, where X is 7 minus the
//	 number of cards in that creature's controller's hand."
//
// X is counted as the spell resolves (CR 608.2h), off the hand of whoever
// controls the creature then. A hand of seven or more makes the
// calculation negative, and CR 107.1b uses zero in its place (the effect
// neither doubles nor sets power and toughness), so it is -0/-0 rather
// than a pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cd99378-e8e2-43fe-80fc-13d680746a4a",
		Name:         "Flunk",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			x := 7 - len(effects.AllHandCardIDs(ctx.Game, controller))
			if x < 0 {
				x = 0
			}
			return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x,
				Label: "Flunk — gets -X/-X until end of turn"}.Apply(ctx)
		},
	})
}
