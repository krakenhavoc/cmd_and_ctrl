package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lava Runner — Creature — Lizard {1}{R}{R}, 2/2 (Mercadian Masques, S82 OLD
// b683, #5619):
//
//	"Haste
//	 Whenever this creature becomes the target of a spell or ability, that
//	 spell or ability's controller sacrifices a land of their choice."
//
// Haste is a printed keyword. The trigger reads the controller of the
// targeting spell or ability off the becomes-target event (its actor), who
// may be the Runner's own controller, and that player picks one of their own
// lands (Crude Bent Blade's edict, PlayerSacrificesForEffect). A player with
// no land to sacrifice sacrifices nothing.
//
// No simplification.
func init() {
	const label = "Lava Runner — that spell or ability's controller sacrifices a land of their choice"
	effects.Register(effects.Spec{
		OracleID:        "48a77c0c-a179-47f2-83b2-da80ba5448f9",
		Name:            "Lava Runner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == source.InstanceID && ev.Actor != uuid.Nil
			}, label, func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Trigger.Event.Actor, effects.SacrificeSpec("a land", effects.Land()), label)
				return nil
			}),
		},
	})
}
