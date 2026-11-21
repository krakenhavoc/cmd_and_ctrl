package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drown in Ichor — Sorcery {1}{B} (Phyrexia: All Will Be One, #5479):
//
//	"Target creature gets -4/-4 until end of turn. Proliferate."
//
// Whisper of the Dross at -4/-4: the shrink lands first, then the player's
// proliferate choice. With its only target gone the spell does not resolve
// at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4912bc3-bee9-4a2f-a13e-3a99018f8a65",
		Name:         "Drown in Ichor",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: -4, Toughness: -4,
				Label: "Drown in Ichor — -4/-4 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
