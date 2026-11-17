package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collective Nightmare — Instant {2}{B} (March of the Machine): "Convoke.
// Target creature gets -3/-3 until end of turn."
func init() {
	Register(Spec{
		OracleID:     "fa7b7db2-c830-405c-a5e6-e5431d573ed3",
		Name:         "Collective Nightmare",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCardIDs(ctx) {
				if err := (BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Collective Nightmare — -3/-3"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
