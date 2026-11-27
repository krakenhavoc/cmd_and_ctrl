package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yawgmoth's Edict — Enchantment {1}{B} (Urza's Saga, S82 OLD batch 732, #5624):
//
//	"Whenever an opponent casts a white spell, that player loses 1 life and
//	 you gain 1 life."
//
// Triggers on an opponent's cast of a spell that is white as it is cast (read
// through the layers and the source's text, CR 612.2); the casting player is
// the one who loses the life, as a life change that names the Edict and no
// damage, and its controller gains 1 whether or not the loss happened. A
// player who has left the game by then loses nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "544eca0d-63cd-4cc5-9318-c2006cdec170",
		Name:         "Yawgmoth's Edict",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.OfColor("W")),
				"Yawgmoth's Edict — that player loses 1 life and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger != nil {
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), item.Trigger.Event.Actor, -1); err != nil {
							return err
						}
					}
					return effects.GainLife{Amount: 1}.Apply(ctx)
				}),
		},
	})
}
