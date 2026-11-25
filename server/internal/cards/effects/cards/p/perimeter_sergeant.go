package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Perimeter Sergeant — Creature — Human Soldier {2}{W}, 3/2 (Ikoria: Lair
// of Behemoths, #5512):
//
//	"Whenever this creature attacks, other Humans you control get +1/+0
//	 until end of turn."
//
// Hamlet Captain's shape: the Humans are picked when the ability resolves
// (CR 611.2c), the Sergeant itself excluded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e1b21f9-c890-4d27-81d0-d38bf7493d33",
		Name:         "Perimeter Sergeant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Perimeter Sergeant — other Humans you control get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					self := ctx.Source()
					match := effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Human"),
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
					return effects.BoostUntilEOT{Match: match, Power: 1, Label: "Perimeter Sergeant — +1/+0"}.Apply(ctx)
				}),
		},
	})
}
