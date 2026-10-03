package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Bodyguard — Creature — Human {3}{W}{W}, 2/5:
//
//	"As long as this creature is untapped, all damage that would be dealt
//	 to you by unblocked creatures is dealt to this creature instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) to the
// Bodyguard, while untapped, of damage from an unblocked attacking
// creature (CR 509.1h). The rulings: a blocked trampler's damage is not
// redirected; it works while the Bodyguard is blocking; with several, you
// choose which receives each event.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d29078c0-1fb8-437a-81d1-bb319f646941",
		Name:         "Veteran Bodyguard",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			staticRedirection("Veteran Bodyguard — damage to you from unblocked creatures is dealt to it instead",
				redirectWhere{applies: untappedAnd(damageToYouFromAnUnblockedCreature(false)), to: toThisPermanent}),
		},
	})
}
