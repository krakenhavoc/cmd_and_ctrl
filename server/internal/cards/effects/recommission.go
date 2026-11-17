package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Recommission — Sorcery {1}{W} (The Brothers' War, #5482):
//
//	"Return target artifact or creature card with mana value 3 or less
//	 from your graveyard to the battlefield. If a creature enters this
//	 way, it enters with an additional +1/+1 counter on it."
//
// An artifact creature is a creature, so it gets the counter; a plain
// artifact does not. The counter is part of the entry (CR 614.1c), seen by
// counter-doubling replacements and already there for the creature's own
// enter triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e43e8c00-f1e7-49d2-bb27-f1c16d85cc22",
		Name:         "Recommission",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target artifact or creature card with mana value 3 or less from your graveyard",
			YouOwn(), Or(Artifact(), Creature()), ManaValueLE(3)),
		Purpose: ForTargets(Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return s82Bro10ReturnTargetToBattlefield(true)(ctx.Game, item)
		},
	})
}
