package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deserter's Disciple — Creature — Human Rebel Ally {1}{R}, 2/2 (Avatar:
// The Last Airbender, #4459):
//
//	"{T}: Another target creature you control with power 2 or less can't
//	 be blocked this turn."
//
// The power is a targeting restriction, so it is checked again as the
// ability resolves: a target pumped to 3 in response makes it fizzle,
// one pumped after it resolved stays unblockable, and one already
// blocked stays blocked (the rulings).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73ca3132-d849-4cf9-beba-75ba09789332",
		Name:         "Deserter's Disciple",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Another target creature you control with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target creature you control with power 2 or less", effects.YouControl(), effects.PowerLE(2))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Deserter's Disciple — can't be blocked",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
