package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "d090131d-428c-4a98-afc5-595b9bc7adce",
		Name:            "Sulu, Space Swashbuckler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondEquipmentOrSpacecraftEntered{},
				effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
					func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
						return effects.ThisBecameTapped(ev, source, lki, g) &&
							effects.TrekB4EquipmentOrSpacecraftEnteredUnderYou(g, source.Controller)
					},
					"Sulu — draw a card",
					effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
