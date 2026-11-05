package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borg Incursion — Instant {2}{W} (Star Trek Commander, #4974):
//
//	"Create two 1/1 colorless Borg artifact creature tokens, then
//	 artifact creatures you control gain protection from the color of
//	 your choice until end of turn."
//
// The tokens come first, so they are among the artifact creatures
// protected. The colour is chosen as the spell resolves and asked as
// the colour to be protected FROM (ColorForProtection, #780); the set
// protected is the artifact creatures you control at that moment
// (CR 611.2c), so one that enters later this turn is not.
func init() {
	Register(Spec{
		OracleID:     "f00dfc16-c253-4483-94fb-3c2beabd5820",
		Name:         "Borg Incursion",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			then := ctx.ItemTail(trekB31ArtifactCreaturesProtectionFromChosenColor)
			then.Words = []string{"Borg Incursion"}
			ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(),
				"Borg Incursion — choose a color", then)
			return nil
		},
	})
}
