package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harsh Annotation — Instant {1}{W} (Secrets of Strixhaven, #4456):
//
//	"Destroy target creature. Its controller creates a 1/1 white and black
//	 Inkling creature token with flying."
//
// Beast Within's order and stale-controller guard
// (socB26DestroyThenItsControllerCreates): the creature's controller is read
// before the destroy, so the Inkling goes to whoever held it when the spell
// resolved, and it is made even if the destroy did nothing (an indestructible
// creature still gives its controller the Inkling). If the target left in
// response, the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1657be64-b38f-4074-a743-bf558dc57901",
		Name:         "Harsh Annotation",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		// #2679: what the target's controller is given back.
		Purpose:   ForTargets(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1})),
		OnResolve: socB26DestroyThenItsControllerCreates("1/1 white and black Inkling with flying"),
	})
}
