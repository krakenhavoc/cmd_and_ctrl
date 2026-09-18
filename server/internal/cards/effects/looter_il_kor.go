package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Looter il-Kor — "Shadow (This creature can block or be blocked by only
// creatures with shadow.) Whenever this creature deals damage to an opponent,
// draw a card, then discard a card."
func init() {
	Register(Spec{
		OracleID:        "c87e3b1d-2d24-4f1c-84e8-1c21a347cdc2",
		Name:            "Looter il-Kor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Source != source.InstanceID || ev.Amount <= 0 {
					return false
				}
				p := g.PlayerByIDForEffect(ev.Target)
				return p != nil && p.ID != source.Controller
			}, "Looter il-Kor — draw, then discard", func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 1, 1)
			}),
		},
	})
}
