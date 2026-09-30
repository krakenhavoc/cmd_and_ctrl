package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murdock's Crusade — {1}{W} Sorcery:
//
//	"Teamwork 4 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 4 or more.)
//	 Choose one. If this spell was cast using teamwork, choose both
//	 instead.
//	 • Street Justice — Exile target creature with toughness 4 or
//	   greater.
//	 • Legal Justice — Exile target enchantment with mana value 4 or
//	   greater."
//
// #1703: Teamwork(4) with InsteadIf(2, TeamworkUsed). The anchor words
// are flavour (CR 207.2c) and ride the bullet labels as printed.
// Toughness is the current, layered toughness, read at announce and
// again at resolution (CR 608.2b). No simplification.
func init() {
	Register(Spec{
		OracleID:      "36330947-5ef9-4bff-b541-bfbddd9715a3",
		Name:          "Murdock's Crusade",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(4)},
		Modes: ChooseOne(
			ModeDoing("Street Justice — Exile target creature with toughness 4 or greater.",
				TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
				exileTheModesTarget),
			ModeDoing("Legal Justice — Exile target enchantment with mana value 4 or greater.",
				TargetPermanent("target enchantment with mana value 4 or greater", Enchantment(), ManaValueGE(4)),
				exileTheModesTarget),
		).InsteadIf(2, TeamworkUsed),
	})
}

// exileTheModesTarget is "exile target <thing>" as a modal bullet's
// body — both of Murdock's Crusade's bullets, which differ only in the
// clause on the option.
func exileTheModesTarget(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ExileTarget{Target: t.ID}.Apply(ctx)
}
