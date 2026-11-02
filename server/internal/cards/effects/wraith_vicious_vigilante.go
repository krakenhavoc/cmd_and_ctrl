package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wraith, Vicious Vigilante — {1}{W}{U} Legendary Creature — Human Detective
// Hero, 1/1 (Marvel's Spider-Man, #4460):
//
//	"Double strike
//	 Fear Gas — Wraith can't be blocked."
//
// "Fear Gas" is an ability word and carries no rules meaning. Blighted
// Agent's RestrictSelf(CantBeBlocked) is the whole second line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "38684c31-e8c6-4825-8740-1a3d34b58b39",
		Name:            "Wraith, Vicious Vigilante",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
