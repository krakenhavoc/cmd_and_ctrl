package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archon's Glory — Instant {W} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Target creature gets +2/+2 until end of turn. If this spell was
//	 bargained, that creature also gains flying and lifelink until end of turn."
//
// Bargain is the shared optional cost (g2Bargain), read back from the payment
// record at resolution as Candy Grapple reads it. The pump and the keywords are
// Academic Ascent's two until-end-of-turn effects on the same target; the
// keywords come only from a bargained cast. The declared purpose is the +2/+2,
// the least the spell does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "bc5631ae-494e-4987-8d17-57674bfbdb9f",
		Name:          "Archon's Glory",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.G2Bargain()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Archon's Glory — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			if !effects.G2WasBargained(ctx) {
				return nil
			}
			return effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"flying", "lifelink"},
				Label:    "Archon's Glory — flying and lifelink",
			}.Apply(ctx)
		},
	})
}
