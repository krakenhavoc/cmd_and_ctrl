package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aya of Alexandria — Legendary Creature — Human Assassin {2}{R}{W}, 4/3
// (Assassin's Creed, #5948):
//
//	"Menace, lifelink
//	 Whenever a historic creature you control deals combat damage to a player,
//	 create a 1/1 black Assassin creature token with menace. (Artifacts,
//	 legendaries, and Sagas are historic.)"
//
// One trigger per historic creature that connects (Ezio, Blade of
// Vengeance's read of the damage event), Aya herself included. "Historic" is
// CR 700.6, Jhoira's b09IsHistoric read off the dealer as it dealt the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27b690f2-d7ce-43b6-8e8b-98a8ffe487bb",
		Name:            "Aya of Alexandria",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && effects.B09IsHistoric(dealer)
			}, "Aya of Alexandria — create a 1/1 black Assassin creature token with menace",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Assassin with menace"), N: 1})),
		},
	})
}
