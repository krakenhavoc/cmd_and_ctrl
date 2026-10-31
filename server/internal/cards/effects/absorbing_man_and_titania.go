package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Absorbing Man and Titania — {3}{R}{G} Legendary Creature — Human
// Villain 4/5 (Marvel Super Heroes Commander, #4455):
//
//	"Double all damage that creature sources you control would deal."
//
// Gratuitous Violence's replacement, on a creature: the source's type
// and controller are read from the damage event's last-known
// characteristics, so a creature that traded is still doubled, and a
// burn spell is not. The shared body is doubleDamageFromCreaturesYouControl.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1afb796-1b0b-4f41-9c8f-5320c9af5989",
		Name:         "Absorbing Man and Titania",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			doubleDamageFromCreaturesYouControl("Absorbing Man and Titania: double the damage"),
		},
	})
}
