package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Lost and the Damned — Enchantment {1}{U}{R} (Warhammer 40,000
// Commander):
//
//	"Whenever a land you control enters from anywhere other than your hand
//	 or you cast a spell from anywhere other than your hand, create a 3/3
//	 red Spawn creature token."
//
// One printed trigger over two events. A land entering reads the zone the
// permanent came from (EventETB.EnteredFrom, with its owner, so a land put
// onto the battlefield from an opponent's hand counts): a land played or
// put from your hand does not, a land from a library, graveyard or exile
// does, and a land token, which comes from no zone, does not. A cast reads
// the zone the spell was cast from (the cast event's OldZone); a spell cast
// with no recorded origin is treated as a hand cast, the answer that cannot
// over-fire. Each qualifying event is its own trigger (CR 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22e70ad0-1a9c-4d9e-8cc1-ebea485282ca",
		Name:         "The Lost and the Damned",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventCast},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					switch ev.Kind {
					case game.EventCast:
						return effects.B25CastFromNotHand(ev, source.Controller)
					case game.EventETB:
						c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
						if !ok || !c.IsLand() || ev.EnteredFrom == "" {
							return false
						}
						return ev.EnteredFrom != game.ZoneHand || ev.EnteredFromOwner != source.Controller
					}
					return false
				},
				"The Lost and the Damned — create a 3/3 red Spawn creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 red Spawn"), N: 1})),
		},
	})
}
