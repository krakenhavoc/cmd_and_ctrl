package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stigma Lasher — Creature — Elemental Shaman {R}{R}, 2/2:
//
//	"Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)
//	 Whenever this creature deals damage to a player, that player can't
//	 gain life for the rest of the game."
//
// Wither rides PrintedKeywords (ADR 0056). The trigger fires on any
// damage the Lasher deals to a player, combat or not, and writes ADR
// 0107 §5's rest-of-the-game grant (CR 119.7, CR 611.2a, #1880): stored
// on the game, so it outlives the Lasher.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7e4a1286-cac6-491f-a32d-64fb98d35de2",
		Name:            "Stigma Lasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && ev.Amount > 0 && g.PlayerByIDForEffect(ev.Target) != nil
			}, "Stigma Lasher — that player can't gain life for the rest of the game", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return PlayerCantGainLifeForRestOfGame{
					Player: item.Trigger.Event.Target,
					Label:  "Stigma Lasher — can't gain life for the rest of the game",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
