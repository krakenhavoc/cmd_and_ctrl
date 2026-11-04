package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sulu, Space Swashbuckler — Legendary Creature — Human Pilot {1}{W},
// 2/2 (Star Trek, #4973):
//
//	"First strike
//	 Whenever Sulu becomes tapped, if an Equipment or Spacecraft entered
//	 the battlefield under your control this turn, draw a card."
//
// "Becomes tapped" is any tap, an attack included (ThisBecameTapped).
// The "if" is an intervening if (CR 603.4): asked as Sulu taps and again
// as the trigger resolves, off the turn's entry tally.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d090131d-428c-4a98-afc5-595b9bc7adce",
		Name:            "Sulu, Space Swashbuckler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ThisBecameTapped(ev, source, lki, g) &&
						trekB4EquipmentOrSpacecraftEnteredUnderYou(g, source.Controller)
				},
				"Sulu — draw a card",
				func(g *game.Game, item *game.StackItem) error {
					if !trekB4EquipmentOrSpacecraftEnteredUnderYou(g, item.Controller) {
						return nil
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
