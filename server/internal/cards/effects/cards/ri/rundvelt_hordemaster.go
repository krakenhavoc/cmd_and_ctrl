package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rundvelt Hordemaster — Creature — Goblin Warrior {1}{R}, 1/1
// (Dominaria United):
//
//	"Other Goblins you control get +1/+1.
//	 Whenever this creature or another Goblin you control dies, exile the
//	 top card of your library. If it's a Goblin creature card, you may cast
//	 that card until the end of your next turn."
//
// The lord is a Layer 7c static over the OTHER Goblins its controller
// controls, read through the layered view (a changeling counts). The dies
// trigger is Pashalik Mons': the source's own death, or another Goblin the
// controller controlled as it last existed on the battlefield (CR 603.10a).
// The top card is exiled whatever it is; only a Goblin creature card is
// given a cast permission, which is the card's own printed filter
// (CardInZone), lasting through the end of its controller's next turn. It
// is a cast, never a land play, and a Goblin creature card that is not cast
// in time stays exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e080c7d-102c-4d28-bcb9-62b36d400d13",
		Name:         "Rundvelt Hordemaster",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CardDied(ev, source) || effects.B25AnotherGoblinYouControlDied(ev, source, g)
			}, "Rundvelt Hordemaster — exile the top card of your library; if it's a Goblin creature card, you may cast it until the end of your next turn",
				rundveltHordemasterExile),
		},
	})
}

// rundveltHordemasterExile exiles the top card; the permission names the
// Goblin creature cards only, so any other card is exiled and stays so.
func rundveltHordemasterExile(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	goblin := effects.NewContext(g, item).TextSubtype("Goblin")
	_, err := g.ExileTopWithPermissionForEffect(controller, controller, 1, game.CastPermission{
		CastOnly: true,
		Duration: g.UntilEndOfYourNextTurnDuration(controller),
		Filter:   game.PermissionFilter{CardInZone: true, CreatureOnly: true, CreatureType: goblin},
	})
	return err
}
