package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Severed Strands — Sorcery {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 You gain life equal to the sacrificed creature's toughness. Destroy
//	 target creature an opponent controls."
//
// The life is the sacrificed creature's toughness as it last existed on
// the battlefield (the 2018-10-05 ruling, CR 608.2h), read off the
// payment record (ADR 0113 §1). The rulings: an illegal target means the
// spell does not resolve and no life is gained (CR 608.2b); a legal
// target that can't be destroyed still gains the life.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "25f0527c-1340-4218-970f-9e4f84ef96e8",
		Name:           "Severed Strands",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GainLife{Amount: sacrificedToughness(ctx)}).Apply(ctx); err != nil {
				return err
			}
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: legal[0].ID}.Apply(ctx)
		},
	})
}
