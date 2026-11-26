package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cult of the Waxing Moon — Creature — Human Shaman {4}{G}, 5/4 (Shadows
// over Innistrad, #5543):
//
//	"Whenever a permanent you control transforms into a non-Human creature,
//	 create a 2/2 green Wolf creature token."
//
// EventTransform is the turn-over (a convert counts too, CR 701.27e). The
// permanent is read as it is now, through the layers: on the battlefield,
// under your control, a creature, and not a Human (a word of this card's
// text, CR 612.2). A permanent that enters on its back face never
// transformed, so it makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ea5409e-9b39-434c-bdb8-8c798128f199",
		Name:         "Cult of the Waxing Moon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTransform, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneBattlefield {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller && c.IsCreature() && !c.HasSubtype(source.TextSubtype("Human"))
			}, "Cult of the Waxing Moon — create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
