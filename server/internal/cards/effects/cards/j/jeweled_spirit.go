package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeweled Spirit — Creature — Spirit {3}{W}{W}, 3/3 (Prophecy):
//
//	"Flying
//	 Sacrifice two lands: This creature gains protection from artifacts or
//	 from the color of your choice until end of turn."
//
// The sacrifice of any two lands (SacrificeN), and a choice made as the
// ability RESOLVES (not a mode chosen on activation, so an opponent responds
// before it is known): artifacts, or a color, asked by Mother of Runes's
// color prompt. Either grants the protection keyword to this creature until
// end of turn, only while it is still the permanent the ability came from
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cddf2e46-79dc-4d6f-bb67-db55d966e8e2",
		Name:            "Jeweled Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice two lands: This creature gains protection from artifacts or from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.SacrificeN(2, "two lands", effects.Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.PickOption{
					Question: "Jeweled Spirit — gain protection from artifacts or from a color?",
					Options: []game.ChoiceOption{
						{Label: "Protection from artifacts"},
						{Label: "Protection from the color of your choice"},
					},
					ThenKey:  jeweledSpiritPicked,
					Carry:    []uuid.UUID{item.SourceCardID},
					WithItem: true,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// jeweledSpiritPicked is the choice: artifacts at once, or the colour prompt.
// Keyed (#4253): an on-disk key, never renamed, never reused.
var jeweledSpiritPicked = effects.OptionPickThen("option-pick/jeweled-spirit-protection", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	spirit := r.Carry[0]
	if r.Index == 0 {
		return effects.GrantKeywordUntilEOT{
			Target:   spirit,
			Keywords: []string{"protection from artifacts"},
			Label:    "Jeweled Spirit — protection from artifacts",
		}.Apply(ctx)
	}
	effects.ChooseProtectionColorFor(ctx, r.Chooser, spirit, "Jeweled Spirit", "Jeweled Spirit — choose a color")
	return nil
})
