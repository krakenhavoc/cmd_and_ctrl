package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazuul's Toll Collector — Creature — Ogre Warrior {2}{R}, 3/2 (Oath of the
// Gatewatch, #5545):
//
//	"{0}: Attach target Equipment you control to this creature. Activate only
//	 as a sorcery."
//
// Not an equip ability: the Equipment is the target and the Collector the
// host (Brass Squire's attach, with the host fixed). The ability costs {0} and
// is sorcery speed. A target that has left, or an Equipment that can't
// legally be attached, does nothing (CR 608.2b, CR 301.5c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b0694ff-aa92-4a2e-a366-69fdf5a57c4b",
		Name:         "Kazuul's Toll Collector",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{0}: Attach target Equipment you control to this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{0}"),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target Equipment you control", effects.HasSubtype("Equipment"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				equip, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return g.AttachForEffect(equip.ID, game.TargetRef{Kind: game.TargetCard, ID: ctx.Source()})
			},
		}},
	})
}
