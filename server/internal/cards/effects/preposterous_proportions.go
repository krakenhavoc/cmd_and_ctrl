package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Preposterous Proportions — Sorcery {5}{G}{G} (Foundations, #4465):
//
//	"Creatures you control get +10/+10 and gain vigilance until end of
//	 turn."
//
// Overrun's shape: the boost and the keyword land on the creatures
// you control as it resolves; one that enters later is not affected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e03aad2-6426-4d18-a0e5-bc3d8e6b7b11",
		Name:         "Preposterous Proportions",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{
				Match:     yours,
				Power:     10,
				Toughness: 10,
				Label:     "Preposterous Proportions — +10/+10",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"vigilance"},
				Label:    "Preposterous Proportions — vigilance",
			}.Apply(ctx)
		},
	})
}
