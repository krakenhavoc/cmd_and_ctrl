package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reach for the Sky — Enchantment — Aura {3}{G} (Outlaws of Thunder
// Junction, #4468):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +3/+2 and has reach.
//	 When this Aura is put into a graveyard from the battlefield, draw a
//	 card."
//
// Silken Strength's static shape (a layer 7c +3/+2 and a layer 6 reach
// grant, both scoped to the enchanted creature) under Rancor's graveyard
// trigger: EventLTB fires only from the battlefield, so "landed in a
// graveyard" (cardDied) is the whole condition. The creature dying, the
// Aura destroyed, or the Aura falling off as a state-based action all
// draw the card; the Aura exiled or bounced does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68930fb6-d833-40cf-8917-1a1bf60edbe3",
		Name:            "Reach for the Sky",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 2),
			effects.GrantToAttached("reach"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source)
			}, "Reach for the Sky — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
