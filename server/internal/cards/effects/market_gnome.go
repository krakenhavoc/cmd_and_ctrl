package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Market Gnome — Artifact Creature — Gnome {W}, 0/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature dies, you gain 1 life and draw a card.
//	 When this creature is exiled from the battlefield while you're
//	 activating a craft ability, you gain 1 life and draw a card."
//
// Two triggers, one printed ability each. The second reads the cause the
// engine stamps on the exile of a craft ability's materials
// (game.MoveCauseCraftCost, ADR 0137: the materials are exiled as the
// ability is activated, CR 702.167a, 602.2b): the Gnome left the
// battlefield for exile, and its controller is the player paying. It goes on
// the stack above the craft ability (the ruling), so the life and the card
// come before the crafted card returns. An exile by a spell or an ability
// is not "while you're activating".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12fa17b2-158e-4d53-897f-81111f912b00",
		Name:         "Market Gnome",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Market Gnome — you gain 1 life and draw a card",
				Do(GainLife{Amount: 1}, DrawCards{N: 1})),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.NewZone == game.ZoneExile &&
					ev.Cause == game.MoveCauseCraftCost && ev.CauseController == source.Controller
			}, "Market Gnome — you gain 1 life and draw a card",
				Do(GainLife{Amount: 1}, DrawCards{N: 1})),
		},
	})
}
