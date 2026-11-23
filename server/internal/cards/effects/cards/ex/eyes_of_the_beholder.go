package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eyes of the Beholder — Instant {4}{B}{B} (Adventures in the Forgotten
// Realms):
//
//	"Target creature gets -11/-11 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fae0474-25ec-474f-8c7f-9e9061835030",
		Name:         "Eyes of the Beholder",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -11, -11)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -11,
				Toughness: -11,
				Label:     "Eyes of the Beholder — -11/-11",
			}.Apply(ctx)
		},
	})
}
