package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terminate — Instant {B}{R} (EDHREC rank 224):
//
//	"Destroy target creature. It can't be regenerated."
//
// Unconditional creature removal at instant speed; Doom Blade
// without the colour clause.
//
// "It can't be regenerated" is a no-op rather than a simplification,
// for the reason pongify.go records: regeneration is not modelled
// anywhere, so there is no shield for the clause to override. If a
// regeneration shield ever lands, this card must be revisited along
// with Pongify and Rapid Hybridization — grep for this sentence.
func init() {
	Register(Spec{
		OracleID: "6257c2fd-005f-41e3-8a72-af76df1eb134",
		Name:     "Terminate",
		Targets:  TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
