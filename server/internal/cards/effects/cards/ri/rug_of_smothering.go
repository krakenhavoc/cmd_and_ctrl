package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rug of Smothering — Artifact Creature — Construct {3}, 1/3 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 Whenever a player casts a spell, they lose 1 life for each spell they've
//	 cast this turn."
//
// Any player's cast, including the Rug's controller's. The number is read as
// the trigger resolves: the spells that player has cast this turn, the one
// that triggered it included. A spell that was countered was still cast and still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da1035c4-bafe-434e-b1bd-5b08960d4c91",
		Name:            "Rug of Smothering",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Rug of Smothering — that player loses 1 life for each spell they've cast this turn", func(g *game.Game, item *game.StackItem) error {
				caster := effects.TriggeringActor(item)
				if caster == uuid.Nil {
					return nil
				}
				n := g.CastTallyFor(caster).Total
				if n <= 0 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, caster, -n)
			}),
		},
	})
}
