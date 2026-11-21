package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lifetap — Enchantment {U}{U}:
//
//	"Whenever a Forest an opponent controls becomes tapped, you gain 1
//	 life."
//
// A "becomes tapped" trigger (EventTapCard; #4534). A land tapped for mana
// emits it too (a land that enters tapped, or an attacking land, is read as
// the engine does). A Forest is read through the object's layered subtypes, so
// a land made a Forest counts and a Forest that stopped being one does not;
// "opponent" is the Forest's controller against Lifetap's controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52ac09af-2aa7-4d80-be26-3e6a6efd5c23",
		Name:         "Lifetap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil || ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand() && c.Controller != source.Controller && c.HasSubtype(source.TextSubtype("Forest"))
			}, "Lifetap — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
