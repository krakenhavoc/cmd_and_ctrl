package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2a99c0d6-3200-44ff-8d4c-f7a38cecb05f",
		Name:         "Fire Nation Occupation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fire Nation Occupation — create a 2/2 red Soldier creature token with firebending 1",
				Do(CreateToken{Template: firebendingSoldierToken(), N: 1})),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !isActivePlayer(g, source.Controller)
			}, "Fire Nation Occupation — create a 2/2 red Soldier creature token with firebending 1 (a spell cast during an opponent's turn)",
				Do(CreateToken{Template: firebendingSoldierToken(), N: 1})),
		},
	})
}
