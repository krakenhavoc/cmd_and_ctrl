package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mardu Devotee — Creature — Human Scout {W}, 1/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"When this creature enters, scry 2. (Look at the top two cards of
//	 your library, then put any number of them on the bottom and the
//	 rest on top in any order.)
//	 {1}: Add {R}, {W}, or {B}. Activate only once each turn."
//
// The entry trigger is the ordinary scry (CR 701.22a). The mana ability
// is Abzan Devotee's: a filter paid with {1} and no tap, limited by the
// declarative OncePerTurn bit, counted per object, so a Devotee that
// leaves and comes back has a fresh use (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24ee57b5-a505-4264-97ca-8b8dfb12298e",
		Name:         "Mardu Devotee",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mardu Devotee — scry 2", Do(Scry{N: 2})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Mana: "{1}"},
			Produced:    "{R|W|B}",
			Label:       "{1}: Add {R}, {W}, or {B}",
			OncePerTurn: true,
		}},
	})
}
