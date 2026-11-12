package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skull Skaab — Creature — Zombie {U}{B}, 2/2:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 Whenever a creature you control exploits a nontoken creature,
//	 create a 2/2 black Zombie creature token."
//
// Two printed abilities, two triggers. Exploit is the shared constructor
// (exploit.go); the Skaab prints no "when this creature exploits a
// creature" of its own, so it passes no payoff body.
//
// The second watches game.EventExploited (#5676, CR 702.110b), emitted
// when an exploit ability's sacrifice has settled. It counts any creature
// its controller controls, the Skaab's own exploit included, and only a
// sacrificed creature that was not a token. Exploiting is a sacrifice,
// so the ability looks back (CR 603.10a): a Skaab that was itself the
// nontoken creature sacrificed, to its own exploit or to another
// creature's, still makes its Zombie.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c73839bf-7500-4fd4-aa2b-e2e2e157784d",
		Name:         "Skull Skaab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Exploit("Skull Skaab", game.BodyRef{}),
			WheneverACreatureYouControlExploitsANontokenCreature("Skull Skaab — create a 2/2 black Zombie creature token",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
		},
	})
}
