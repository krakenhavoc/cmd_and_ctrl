package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inner Calm, Outer Strength — Instant — Arcane {2}{G} (Saviors of
// Kamigawa, #5601):
//
//	"Target creature gets +X/+X until end of turn, where X is the number
//	 of cards in your hand."
//
// X is counted once, as the spell resolves (CR 608.2h), and does not
// change if the hand does later in the turn. A target that left in
// response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b7bdbae5-549f-403b-83ca-4f9a1ac93e93",
		Name:         "Inner Calm, Outer Strength",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			n := sokB1CardsInYourHand(ctx)
			return BoostUntilEOT{Target: target, Power: n, Toughness: n, Label: "Inner Calm, Outer Strength — +X/+X until end of turn"}.Apply(ctx)
		},
	})
}
