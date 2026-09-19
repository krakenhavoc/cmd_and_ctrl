package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rule of Law — {2}{W} Enchantment: "Each player can't cast more than
// one spell each turn."
//
// #760's headline card and the simplest possible cast restriction
// (CR 101.2): one predicate over a tally the engine has kept since
// S19. Nothing is stored on the players — the restriction is DERIVED
// from this permanent being on the battlefield, so it lifts the
// moment Rule of Law leaves, and a second copy changes nothing.
//
// "Each player" includes its own controller. Rule of Law is
// symmetrical and the asymmetrical cards say so out loud ("your
// opponents can't…"), which is why OpponentsCantCast is a separate
// constructor rather than a flag on this one.
//
// The tally is bumped AFTER a cast succeeds, so at announce a player
// who has already cast one spell reads Total == 1 and this is their
// second. A player who has cast none reads 0 and is fine.
func init() {
	Register(Spec{
		OracleID:     "53e88e64-6f82-4154-a66e-6aeb0154b368",
		Name:         "Rule of Law",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			EachPlayerMaxSpellsPerTurn(1,
				"Rule of Law — each player can't cast more than one spell each turn."),
		},
	})
}
