package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// You Come to a River — Instant {1}{U} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Fight the Current — Return target nonland permanent to its owner's hand.
//	 • Find a Crossing — Target creature gets +1/+0 until end of turn and can't
//	   be blocked this turn."
//
// The bounce is any player's nonland permanent. The second bullet is one
// target and two until-end-of-turn effects, both pinned to it as the spell
// resolves (CR 611.2c): the pump and Artful Dodge's block restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9913da11-07f0-494a-8c47-11629dcb5be2",
		Name:         "You Come to a River",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Fight the Current — Return target nonland permanent to its owner's hand.",
				effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.BounceToHand{Target: id}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Find a Crossing — Target creature gets +1/+0 until end of turn and can't be blocked this turn.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "You Come to a River — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "You Come to a River — can't be blocked"}.Apply(ctx)
				})), effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 1, 0))),
		),
	})
}
