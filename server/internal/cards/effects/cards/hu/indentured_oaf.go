package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indentured Oaf — Creature — Ogre Warrior {3}{R}, 4/3 (Ravnica: City of Guilds):
//
//	"Prevent all damage that this creature would deal to red creatures."
//
// A prevention replacement (CR 615.1a) over damage the Oaf is the source of, to a
// creature that is red as the damage would be dealt, combat or not. It does not
// touch damage dealt to players, to other colors, or by other sources.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d82bde29-1576-4736-a755-df9c37765202",
		Name:         "Indentured Oaf",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.OldB580PreventDamageFromThisToRedCreatures("Indentured Oaf — prevent all damage this creature would deal to red creatures"),
		},
	})
}
