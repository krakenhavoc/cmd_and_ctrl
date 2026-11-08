package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon's Glory — Instant {W} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Target creature gets +2/+2 until end of turn. If this spell was
//	 bargained, that creature also gains flying and lifelink until end of
//	 turn."
//
// Bargain is the shared optional cost (g2Bargain), read back from the
// payment record at resolution as Candy Grapple reads it. The pump is an
// until-end-of-turn modify on the target; a bargained cast also grants
// flying and lifelink to the same creature. A target gone by resolution
// gets nothing (CR 608.2b). The declared purpose is the unbargained
// +2/+2, the least the spell does.
//
// No simplification.
func init() {
	const label = "Archon's Glory — +2/+2 until end of turn"
	Register(Spec{
		OracleID:      "bc5631ae-494e-4987-8d17-57674bfbdb9f",
		Name:          "Archon's Glory",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: label}).Apply(ctx); err != nil {
				return err
			}
			if !g2WasBargained(ctx) {
				return nil
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "lifelink"}, Label: "Archon's Glory — flying and lifelink until end of turn"}.Apply(ctx)
		},
	})
}
