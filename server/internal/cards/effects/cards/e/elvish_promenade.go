package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Promenade — Kindred Sorcery — Elf {3}{G} (EDHREC rank
// 3068):
//
//	"Create a 1/1 green Elf Warrior creature token for each Elf you
//	 control."
//
// The Elf deck's doubler. Counted at resolution over the permanents
// the caster controls with the Elf subtype (post-layer, so a
// changeling counts); the spell itself is an Elf but is on the
// stack, not under anyone's control, and does not count itself —
// as printed. Zero Elves makes nothing. The tokens are the shared
// Elf Warrior template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7df3e379-c217-416e-a1c8-46338608c49e",
		Name:         "Elvish Promenade",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.B26PermanentsOfSubtypeControlled(ctx.Game, item.Controller, ctx.TextSubtype("Elf"))
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Elf Warrior"), N: n}.Apply(ctx)
		},
	})
}
