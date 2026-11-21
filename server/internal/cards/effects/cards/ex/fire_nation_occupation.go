package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Occupation — Enchantment {2}{B} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"When this enchantment enters, create a 2/2 red Soldier creature
//	 token with firebending 1. (Whenever it attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 Whenever you cast a spell during an opponent's turn, create a 2/2
//	 red Soldier creature token with firebending 1."
//
// Firebender Ascension's Soldier (printedFirebendingSoldierToken). The
// cast trigger is Brineborn Cutthroat's condition: you cast the spell
// while you are not the active player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a99c0d6-3200-44ff-8d4c-f7a38cecb05f",
		Name:         "Fire Nation Occupation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fire Nation Occupation — create a 2/2 red Soldier creature token with firebending 1",
				effects.Do(effects.CreateToken{Template: effects.FirebendingSoldierToken(), N: 1})),
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !effects.IsActivePlayer(g, source.Controller)
			}, "Fire Nation Occupation — create a 2/2 red Soldier creature token with firebending 1 (a spell cast during an opponent's turn)",
				effects.Do(effects.CreateToken{Template: effects.FirebendingSoldierToken(), N: 1})),
		},
	})
}
