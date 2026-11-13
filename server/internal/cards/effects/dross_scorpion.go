package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dross Scorpion — Artifact Creature — Scorpion {4}, 3/1 (Mirrodin,
// #5606):
//
//	"Whenever this creature or another artifact creature dies, you may
//	 untap target artifact."
//
// One trigger with two conditions: the Scorpion's own death, and the
// death of any other artifact creature, anyone's. Both are dies
// triggers, so they look back at the creature as it left the
// battlefield (CR 603.10a): an artifact creature that stopped being an
// artifact before it died does not count, and an animated artifact
// that dies does. The target is chosen as the trigger goes on the stack
// (CR 603.3d) and the "you may" is asked as it resolves; Voltaic Key's
// untap does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "132786bb-157c-4e2c-8f82-13df6a667448",
		Name:         "Dross Scorpion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if ThisDied(ev, source, lki, g) {
					return true
				}
				dead, ok := diedCreature(ev, g)
				return ok && leftAsType(ev, dead, "artifact")
			}, "Dross Scorpion — untap target artifact", untapFirstLegalTarget),
				TargetPermanent("target artifact", Artifact())),
				"Untap target artifact with Dross Scorpion?"),
		},
	})
}
