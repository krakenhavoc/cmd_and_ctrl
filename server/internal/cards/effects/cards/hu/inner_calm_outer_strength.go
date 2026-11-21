package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "b7bdbae5-549f-403b-83ca-4f9a1ac93e93",
		Name:         "Inner Calm, Outer Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			n := effects.SokB1CardsInYourHand(ctx)
			return effects.BoostUntilEOT{Target: target, Power: n, Toughness: n, Label: "Inner Calm, Outer Strength — +X/+X until end of turn"}.Apply(ctx)
		},
	})
}
