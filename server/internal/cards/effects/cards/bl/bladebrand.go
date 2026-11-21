package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladebrand — Instant {1}{B} (Innistrad: Midnight Hunt Commander reprint,
// S82):
//
//	"Target creature gains deathtouch until end of turn.
//	 Draw a card."
//
// The keyword is a layer 6 grant until cleanup (CR 514.2). The draw is not
// part of the target's fate: with the target gone the spell is countered on
// resolution by rules (CR 608.2b), and the draw goes with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9cf9f1e-0280-43b6-9d58-262052dcb8a8",
		Name:         "Bladebrand",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: item.Targets[0].ID, Keywords: []string{"deathtouch"},
				Label: "Bladebrand — deathtouch"}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
