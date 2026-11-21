package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kruge, Genesis Seeker — Legendary Creature — Klingon Officer {2}{R}, 2/2
// (Star Trek Commander, #4974):
//
//	"Whenever Kruge attacks, attacking Klingons you control get +1/+1
//	 until end of turn."
//
// Blaring Captain's attack trigger, narrowed to Klingons you control.
// The set of attacking Klingons is locked in as the trigger resolves
// (CR 611.2c): Kruge himself if he is still a Klingon, and not a Klingon
// put onto the battlefield attacking afterwards. "Klingons" is a word of
// the source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61cef71a-d38b-4b6f-8401-aa4fc861bf52",
		Name:         "Kruge, Genesis Seeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kruge, Genesis Seeker — attacking Klingons you control get +1/+1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.AttackingCreature(), effects.YouControl(), effects.OfSubtype(ctx.TextSubtype("Klingon"))),
						Power:     1,
						Toughness: 1,
						Label:     "Kruge, Genesis Seeker — +1/+1",
					}.Apply(ctx)
				}),
		},
	})
}
