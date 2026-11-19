package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glory Bearers — {3}{W} Enchantment Creature — Human Cleric 3/4:
//
//	"Whenever another creature you control attacks, it gets +0/+1 until
//	end of turn."
//
// One trigger per attacking creature (each creature's attack is its own
// event); the pump goes on the creature the event names.
func init() {
	Register(Spec{
		OracleID:     "8f446660-f72a-454f-bc96-3b6dd8633228",
		Name:         "Glory Bearers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && ev.CardID != source.InstanceID
			}, "Glory Bearers — the attacking creature gets +0/+1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return BoostUntilEOT{
					Target:    item.Trigger.Event.CardID,
					Toughness: 1,
					Label:     "Glory Bearers — +0/+1",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
