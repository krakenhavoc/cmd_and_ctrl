package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planar Void — Enchantment {B}:
//
//	"Whenever another card is put into a graveyard from anywhere, exile that
//	 card."
//
// Any card (not a token) going to any graveyard from the battlefield, hand,
// library, stack or exile, but not Planar Void itself and not a card already in a
// graveyard. It is exiled only if it is still in that graveyard as the trigger
// resolves, so a card another effect moved on in response is left alone.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "994e27e1-0bff-47c9-a1e6-b9e3ae4ffb1e",
		Name:         "Planar Void",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.CardID != source.InstanceID && effects.OldB720CardPutIntoAGraveyard(ev, g)
				}, "Planar Void — exile that card", func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.CardID
					z := g.FindCardZoneForEffect(id)
					if z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.ExileTarget{Target: id}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
