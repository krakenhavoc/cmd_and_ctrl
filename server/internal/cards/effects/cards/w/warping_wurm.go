package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warping Wurm — Creature — Wurm {2}{G}{U}, 1/1 (Mirage, S82 OLD batch 812,
// #5633):
//
//	"Phasing (This phases in or out before you untap during each of your untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)
//	 At the beginning of your upkeep, this creature phases out unless you pay
//	 {2}{G}{U}.
//	 Whenever this creature phases in, put a +1/+1 counter on it."
//
// Phasing is the engine's own keyword. The upkeep trigger is Spindrift
// Drake's pay-or-else with the decline phasing the Wurm out; the phase-in
// trigger is read off the permanent itself (CR 603.10b: Teferi's Imp's seam).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c54c1517-cf7c-4afb-92f8-70c9d1591b45",
		Name:            "Warping Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPhasing},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Warping Wurm — this creature phases out unless you pay {2}{G}{U}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{2}{G}{U}",
						Question: "Warping Wurm — pay {2}{G}{U} or Warping Wurm phases out?",
						Else:     effects.S82OldB812PhaseOutThisUnlessPaid,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WheneverThisPhasesIn("Warping Wurm — put a +1/+1 counter on it", effects.PlusOneCountersOnThis(1)),
		},
	})
}
