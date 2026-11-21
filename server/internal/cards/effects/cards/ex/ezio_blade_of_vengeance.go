package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ezio, Blade of Vengeance — Legendary Creature — Human Assassin
// {3}{U}{B}, 5/5 (Assassin's Creed, #5466):
//
//	"Deathtouch
//	 Whenever an Assassin you control deals combat damage to a player,
//	 draw a card."
//
// One trigger per Assassin that connects (Ezio himself included), read
// off the combat damage event the way Etrata reads its Assassins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2127a245-d664-45cb-8ed1-0024de1a10df",
		Name:            "Ezio, Blade of Vengeance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				src, ok := g.LookupCardForEffect(ev.Source)
				return ok && src.HasSubtype(source.TextSubtype("Assassin"))
			}, "Ezio, Blade of Vengeance — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
