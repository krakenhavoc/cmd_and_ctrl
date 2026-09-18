package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ice Tunnel — Snow Land — Island Swamp (EDHREC rank 4231):
//
//	"({T}: Add {U} or {B}.)
//	 This land enters tapped."
//
// The Dimir member of Kaldheim's snow-dual cycle: a tapped dual whose
// only upside over the common guildgates is the snow supertype, which
// is what a Dead of Winter or an Into the Snow is counting.
//
// The snow supertype rides the printed type line straight off the
// deck importer, so nothing here declares it — the two cards in the
// catalog that count snow permanents read it the same way they read
// any other supertype. What has to be declared is the mana, for the
// reason every nonbasic dual declares it: two basic land types on a
// nonbasic land produce nothing on their own in this engine.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "40c5d6fe-854a-436d-9f80-13eb5f1f8f68",
		Name:          "Ice Tunnel",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "B")},
	})
}
