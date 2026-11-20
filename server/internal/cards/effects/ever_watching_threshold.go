package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ever-Watching Threshold — Enchantment {2}{U} (Commander 2018, #5526):
//
//	"Whenever an opponent attacks, if they attacked you and/or a
//	 planeswalker you control, draw a card."
//
// One trigger per attack declaration (OncePerBatch keeps the first matching
// attacker's event and declines the rest of the batch), and only when at
// least one of the opponent's attackers is attacking you or a planeswalker
// you control (Marchesa's Decree's attacksYouOrYourPlaneswalker; an attack
// on a battle you protect does not count, #4864). A second combat phase is
// a new declaration and can trigger it again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6639b52-b51d-46e7-a83c-c4621b1cecaa",
		Name:         "Ever-Watching Threshold",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && ev.Actor != source.Controller &&
					attacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
			}, "Ever-Watching Threshold — an opponent attacked you or a planeswalker you control: draw a card",
				Do(DrawCards{N: 1}))),
		},
	})
}
