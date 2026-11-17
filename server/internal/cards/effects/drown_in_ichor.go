package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e4912bc3-bee9-4a2f-a13e-3a99018f8a65",
		Name:         "Drown in Ichor",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, -4)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (BoostUntilEOT{Target: item.Targets[0].ID, Power: -4, Toughness: -4,
				Label: "Drown in Ichor — -4/-4 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}
