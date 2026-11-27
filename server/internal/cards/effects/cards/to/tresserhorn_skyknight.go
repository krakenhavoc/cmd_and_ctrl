package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tresserhorn Skyknight — Creature — Zombie Knight {5}{B}{B}, 5/3 (Coldsnap,
// #5597):
//
//	"Flying
//	 Prevent all damage that would be dealt to this creature by creatures with
//	 first strike."
//
// A standing CR 615 prevention of damage to this creature whose source was a
// creature with first strike as the damage would be dealt (last-known
// information, CR 608.2h). Damage from a creature without first strike, or from a
// noncreature source, is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "df344469-1177-4168-91af-3621b2a65182",
		Name:            "Tresserhorn Skyknight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{effects.S82OldB568PreventDamageToThisFromFirstStrikers(
			"Tresserhorn Skyknight — prevent all damage that would be dealt to it by creatures with first strike")},
	})
}
