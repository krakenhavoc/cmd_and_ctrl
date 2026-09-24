package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sudden Shock — Instant {1}{R}:
//
//	"Split second (As long as this spell is on the stack, players
//	 can't cast spells or activate abilities that aren't mana
//	 abilities.)
//	 Sudden Shock deals 2 damage to any target."
//
// Shock with split second (#1519): the classic answer to a creature
// its controller could otherwise save by sacrificing it, regenerating
// it or pumping it in response. The keyword is the whole difference
// from shock.go, and the engine does all of it — the cast path reads
// the declaration below and shuts every non-mana response until this
// leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b139231-dcd2-4b03-bcf2-fcb040617b69",
		Name:            "Sudden Shock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b10DamageEachLegalTarget(ctx, 2)
		},
	})
}
