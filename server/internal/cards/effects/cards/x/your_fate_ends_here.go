package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Your Fate Ends Here — Instant {2}{W} (Reality Fracture, tracker #2795):
//
//	"Destroy target creature or planeswalker with mana value 3 or greater.
//	 Surveil 1. (Look at the top card of your library. You may put it into
//	 your graveyard.)"
//
// A token has mana value 0 and so is never a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28a362c2-0c99-48e0-b9cb-485a6c1f350b",
		Name:         "Your Fate Ends Here",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature or planeswalker with mana value 3 or greater",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueGE(3))),
		OnResolve: effects.DestroyTargetThenSurveil(1),
	})
}
