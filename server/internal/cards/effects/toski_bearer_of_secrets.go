package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toski, Bearer of Secrets — Legendary Creature — Squirrel {3}{G},
// 1/1:
//
//	"This spell can't be countered.
//	 Indestructible
//	 Toski attacks each combat if able.
//	 Whenever a creature you control deals combat damage to a player,
//	 draw a card."
//
// Four printed lines, four pieces of engine vocabulary already on the
// shelf: `Spec.CantBeCountered` (Dragonlord Dromoka), `PrintedKeywords`
// for indestructible, `AttacksEachCombat()` — CR 508.1d's requirement,
// the same one Zurgo Helmsmasher prints on itself — and
// `combatDamageToPlayerBy`, the Edric-shaped trigger for "a creature
// you control dealt combat damage to a player". The last one fires
// once per dealing creature, exactly as printed (no "one or more"
// batching on this card).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a8e707ec-ce77-4bc5-8c76-5ea3e81e8c7f",
		Name:            "Toski, Bearer of Secrets",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Toski, Bearer of Secrets — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
