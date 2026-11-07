package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampaging Spiketail — Creature — Dinosaur {4}{B}{B}, 5/6 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature enters, target creature you control gets +2/+0
//	 and gains indestructible until end of turn.
//	 Swampcycling {2}"
//
// The enters trigger targets as it goes on the stack (CR 603.3d), and the
// Spiketail itself is a legal choice. The +2/+0 is a layer 7c change and
// the indestructible a layer 6 grant, both ending in the cleanup step
// (CR 514.2); a target that is gone by resolution gets nothing (CR
// 608.2b). Swampcycling is Canyon Crawler's typecycling from the hand
// (CR 702.29e).
//
// No simplification.
func init() {
	const label = "Rampaging Spiketail — target creature you control gets +2/+0 and gains indestructible until end of turn"
	Register(Spec{
		OracleID:     "41f6d6d9-3fe5-44b0-904d-1cb90d36f2b6",
		Name:         "Rampaging Spiketail",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, pumpAndGrantFirstTarget(2, 0, label, "indestructible")),
				targetCreatureYouControl()),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
