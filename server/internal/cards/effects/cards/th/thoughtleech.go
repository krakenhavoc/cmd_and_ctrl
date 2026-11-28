package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thoughtleech — Enchantment {G}{G} (Ice Age, S82 OLD batch 851, #5636):
//
//	"Whenever an Island an opponent controls becomes tapped, you may gain 1
//	 life."
//
// Lifetap's "becomes tapped" trigger (EventTapCard) over an Island instead of a
// Forest, and "you may gain 1 life" as Angel's Feather has it: gaining life can
// only help, so the trigger just gains it. The Island is read through the
// object's layered subtypes and "opponent" is its controller against the
// Thoughtleech's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b74d3768-74bb-413c-b7c7-41028ff48225",
		Name:         "Thoughtleech",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil || ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand() && c.Controller != source.Controller && c.HasSubtype(source.TextSubtype("Island"))
			}, "Thoughtleech — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
