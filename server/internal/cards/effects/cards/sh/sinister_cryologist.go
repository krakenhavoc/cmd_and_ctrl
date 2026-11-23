package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinister Cryologist — Creature — Jellyfish Wizard {2}{U}, 2/3 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, target creature an opponent controls gets
//	 -3/-0 until end of turn.
//	 Warp {U}"
//
// Cogwork Wrestler's enters trigger with -3/-0: it targets as it goes on
// the stack and is removed with no legal target (CR 603.3d), and the
// shrink is locked to the target at resolution (CR 611.2c) and ends in
// the cleanup step (CR 514.2). A target no longer legal is left alone
// (CR 608.2b). Warp is the shared alternative cost (CR 702.185), so a
// warped Cryologist still shrinks as it enters and is exiled at the next
// end step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "be3038ff-b39e-4d2e-97ee-85de074ccfd3",
		Name:             "Sinister Cryologist",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{U}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Sinister Cryologist — target creature an opponent controls gets -3/-0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{
						Target: id,
						Power:  -3,
						Label:  "Sinister Cryologist — -3/-0 until end of turn",
					}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -3, 0))),
		},
	})
}
