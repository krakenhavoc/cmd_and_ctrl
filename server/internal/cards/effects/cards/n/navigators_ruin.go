package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Navigator's Ruin — Enchantment {2}{U} (Ixalan):
//
//	"Raid — At the beginning of your end step, if you attacked this turn,
//	 target opponent mills four cards."
//
// Perforating Artist's raid end-step trigger (an intervening "if", CR 603.4)
// with a targeted mill: the still-legal opponent puts the top four cards of
// their library into their graveyard through the mill primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd830c46-3444-488d-ae5b-0a2533158a13",
		Name:         "Navigator's Ruin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OldB223RaidAtYourEndStep("Navigator's Ruin — raid: target opponent mills four cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					p, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.MillCards{Player: p, N: 4}.Apply(ctx)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
