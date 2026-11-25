package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liesa, Shroud of Dusk — Legendary Creature — Angel {2}{W}{W}{B}, 5/5
// (Commander Legends, #6151):
//
//	"Rather than pay {2} for each previous time you've cast this spell from
//	 the command zone this game, pay 2 life that many times.
//	 Flying, lifelink
//	 Whenever a player casts a spell, they lose 2 life."
//
// The first ability replaces the commander tax (CR 903.8): the same count
// of previous casts charges 2 life each instead of {2}
// (Spec.CommanderTaxInLife), checked against the caster's life with the
// cast's other life costs and paid as the spell goes on the stack. The
// trigger hits every caster, Liesa's controller included, and Liesa's own
// cast does not see it (she is not on the battlefield yet).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "2270daaf-252d-40c4-bed5-bd46ae15f041",
		Name:               "Liesa, Shroud of Dusk",
		Completeness:       effects.CompletenessFull,
		PrintedKeywords:    []string{"flying", "lifelink"},
		CommanderTaxInLife: true,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Liesa, Shroud of Dusk — that player loses 2 life", func(g *game.Game, item *game.StackItem) error {
				caster := effects.TriggeringActor(item)
				if caster == uuid.Nil || g.PlayerByIDForEffect(caster) == nil {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, caster, -2)
			}),
		},
	})
}
