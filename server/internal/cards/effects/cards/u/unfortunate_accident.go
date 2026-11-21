package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unfortunate Accident — Instant {B} (Outlaws of Thunder Junction,
// #4468):
//
//	"Spree (Choose one or more additional costs.)
//	 + {2}{B} — Destroy target creature.
//	 + {1} — Create a 1/1 red Mercenary creature token with '{T}: Target
//	   creature you control gets +1/+0 until end of turn. Activate only
//	   as a sorcery.'"
//
// Phantom Interference's Spree (CR 702.172a): two bullets, each with its
// own cost, and only the first targets. The destroy is the shared
// DestroyTheModesTarget; the token is the set's Mercenary template. With
// both bullets chosen and the creature gone by resolution, the destroy
// does nothing (CR 608.2b) and the Mercenary is still made. The second
// bullet's label leaves out the token's quoted ability, which the token
// itself carries: the bullet makes the token and changes no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "011558f3-3c23-430e-a08b-8f91238e7971",
		Name:         "Unfortunate Accident",
		Completeness: effects.CompletenessFull,
		Modes: effects.Spree(
			effects.SpreeModeDoing("Destroy target creature.", "{2}{B}",
				effects.TargetCreature("target creature"),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Create a 1/1 red Mercenary creature token.", "{1}", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.MercenaryToken(), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		),
	})
}
