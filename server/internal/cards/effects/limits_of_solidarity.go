package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Limits of Solidarity — Sorcery {3}{R}
// (Amonkhet, #5535):
//
//	"Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The threaten body Take for a Ride shares, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2184c37-ee92-40f2-b7d1-97c39ce5f277",
		Name:         "Limits of Solidarity",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			return stdB31Threaten(ctx, t, "Limits of Solidarity")
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
