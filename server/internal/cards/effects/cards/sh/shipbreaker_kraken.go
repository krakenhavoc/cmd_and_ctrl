package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shipbreaker Kraken — Creature — Kraken {4}{U}{U}, 6/6 (Commander Legends):
//
//	"{6}{U}{U}: Monstrosity 4. (If this creature isn't monstrous, put four
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, tap up to four target creatures.
//	 Those creatures don't untap during their controllers' untap steps for as
//	 long as you control this creature."
//
// Tidebinder Mage's hold over up to four chosen creatures: each still-legal
// target is tapped, and the hold is recorded only if the Kraken is still under
// your control (CR 611.2b).
//
// No simplification.
func init() {
	t := effects.WhenBecomesMonstrous("Shipbreaker Kraken — tap up to four target creatures; they don't untap while you control Shipbreaker Kraken",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			return effects.TapAndHoldWhileYouControlThis(ctx, effects.HoldTargetIDs(ctx))
		})
	t.Targets = effects.Harmful(effects.TargetCreature("up to four target creatures").WithCount(0, 4))
	effects.Register(effects.Spec{
		OracleID:     "dda7a088-61e3-4c47-b9e0-529fc047f77e",
		Name:         "Shipbreaker Kraken",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{6}{U}{U}"), 4)},
		Triggered:    []game.TriggeredAbility{t},
	})
}
