package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reverberate — Instant {R}{R}:
//
//	"Copy target instant or sorcery spell. You may choose new
//	 targets for the copy."
//
// The plainest CR 706.10 card there is, and the reason the
// primitive exists. Note the target clause has no "you control":
// Reverberate's whole job in Commander is to answer the table's
// best spell by taking a second one for yourself — copy the
// opponent's Time Warp, copy the Expropriate, copy the tutor.
//
// The copy is controlled by Reverberate's controller, not by the
// controller of the spell it copied (CR 706.10a). That asymmetry is
// carried by CopySpell.Controller defaulting to ctx.Controller().
func init() {
	Register(Spec{
		OracleID: "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca",
		Name:     "Reverberate",
		Targets:  instantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CopySpell{
				StackID:          item.Targets[0].ID,
				Controller:       ctx.Controller(),
				ChooseNewTargets: true,
			}.Apply(ctx)
		},
	})
}
