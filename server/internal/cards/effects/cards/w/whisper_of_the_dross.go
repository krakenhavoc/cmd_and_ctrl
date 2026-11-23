package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whisper of the Dross — Instant {B} (Phyrexia: All Will Be One, #5479):
//
//	"Target creature gets -1/-1 until end of turn. Proliferate."
//
// The shrink lands first, then the player's proliferate choice, so a -1/-1
// counter-free creature that just got smaller is not a candidate for
// anything it does not have. With its only target gone the spell does not
// resolve at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23c2cdc9-2b16-423a-b932-2daadeb1021b",
		Name:         "Whisper of the Dross",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: -1, Toughness: -1,
				Label: "Whisper of the Dross — -1/-1 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
