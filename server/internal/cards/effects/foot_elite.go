package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foot Elite — {2}{W/B} Creature — Human Ninja 2/4 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Whenever this creature attacks, another target creature you control
//	 gets +1/+0 and gains indestructible until end of turn."
//
// Gruul Beastmaster's "another target creature you control" attack
// trigger, with the shared pumpAndGrantFirstTarget body: the still-legal
// target (CR 608.2b) gets +1/+0 and indestructible, nothing if it left.
//
// No simplification.
func init() {
	const label = "Foot Elite — +1/+0 and indestructible until end of turn"
	Register(Spec{
		OracleID:     "b6244b6c-664d-417a-bd58-ace5c7a995a2",
		Name:         "Foot Elite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(label, pumpAndGrantFirstTarget(1, 0, label, "indestructible")),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
