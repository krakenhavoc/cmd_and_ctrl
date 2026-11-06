package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omenport Vigilante — Creature — Human Mercenary {1}{W}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"This creature has double strike as long as you've committed a crime
//	 this turn. (Targeting opponents, anything they control, and/or cards
//	 in their graveyards is a crime.)"
//
// A layer 6 keyword grant (CR 613.1f) read live (CR 611.3a) off this
// turn's event log (YouCommittedACrimeThisTurn): it switches on the moment
// the controller announces a crime, stays on for the rest of the turn even
// if that spell is countered, and is off again on the next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "852e3d89-64a0-4b32-be2b-b1d6cd7c2f64",
		Name:         "Omenport Vigilante",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfCreature(target, g, source) && YouCommittedACrimeThisTurn(g, source.Controller)
			}, "double strike"),
		},
	})
}
