package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruthless Cullblade — Creature — Vampire Warrior {1}{B}, 2/1 (Worldwake, #5582):
//
//	"This creature gets +2/+1 as long as an opponent has 10 or less life."
//
// Bloodghast's condition (SelfWhileAnOpponentsLifeAtMost) on a
// layer-7c pump; LifeGatedPump declares the life-total dependency so the
// pump follows a life change at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b7748ad6-a8cd-45e7-a4d6-4ee747722657",
		Name:         "Ruthless Cullblade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{ruthlessCullbladePump()},
	})
}

func ruthlessCullbladePump() game.StaticAbility {
	ab := LifeGatedPump(SelfWhileAnOpponentsLifeAtMost(10), 2, 1)
	ab.Label = "This creature gets +2/+1 as long as an opponent has 10 or less life."
	return ab
}
