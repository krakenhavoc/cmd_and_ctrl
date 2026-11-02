package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Web of Life and Destiny — Enchantment {6}{G}{G} (Marvel's Spider-Man, #4460):
//
//	"Convoke
//	 At the beginning of combat on your turn, look at the top five cards of
//	 your library. You may put a creature card from among them onto the
//	 battlefield. Put the rest on the bottom of your library in a random
//	 order."
//
// Convoke is the spell's tap cost (CR 702.51). The trigger is the shared
// look-and-put primitive, which leaves the rest on the bottom at random
// whether or not a creature was chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33d556c5-1bb4-49bb-88c2-627311eb8c9b",
		Name:         "Web of Life and Destiny",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Web of Life and Destiny — look at the top five cards, you may put a creature card onto the battlefield",
				LookAtTopThenMayPutOntoBattlefield(5, Creature(), 1,
					"Web of Life and Destiny — put a creature card onto the battlefield")),
		},
	})
}
