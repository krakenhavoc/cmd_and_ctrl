package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathgazer — Creature — Lizard {3}{B}, 2/2 (Eighth Edition, #5607):
//
//	"Whenever this creature blocks or becomes blocked by a nonblack
//	 creature, destroy that creature at end of combat."
//
// Corrosive Ooze's shape: the block event names both creatures, so one
// trigger covers both halves. The other creature's colour is read as the
// trigger fires, and the destruction is scheduled for the beginning of the
// end of combat step (CR 511.2); a creature that has left by then is not
// destroyed (CR 400.7).
//
// No simplification.
func init() {
	const label = "Deathgazer — destroy that creature at end of combat"
	effects.Register(effects.Spec{
		OracleID:     "0f772f35-6d98-4ac8-88e5-02539737078b",
		Name:         "Deathgazer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.S82OldB614FoeIsNonblack, label, effects.S82OldB614DestroyTheFoeAtEndOfCombat(label)),
		},
	})
}
