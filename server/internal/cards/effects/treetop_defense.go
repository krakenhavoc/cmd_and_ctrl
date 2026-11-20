package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treetop Defense — Instant {1}{G} (Portal, #5630):
//
//	"Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Creatures you control gain reach until end of turn. (They can block creatures with flying.)"
//
// Heavy Fog's cast restriction (youWereAttackedThisStep): the declare
// attackers step, in which a creature was declared attacking you.
// Reach goes to the creatures you control as it resolves; one that
// arrives later doesn't gain it (CR 611.2c). Reach lets a creature
// block a creature with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "f215d0f9-a53e-431a-a70d-9dc4e3caa41e",
		Name:               "Treetop Defense",
		Completeness:       CompletenessFull,
		CastCondition:      youWereAttackedThisStep,
		CastConditionLabel: attackedThisStepLabel,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"reach"}, Label: "Treetop Defense — reach"}.Apply(ctx)
		},
	})
}
