package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eight-and-a-Half-Tails — Legendary Creature — Fox Cleric {W}{W}, 2/2
// (Champions of Kamigawa, S82 GAP2 b05 #5948, #6088):
//
//	"{1}{W}: Target permanent you control gains protection from white
//	 until end of turn.
//	 {1}: Target spell or permanent becomes white until end of turn."
//
// The first ability grants the keyword until end of turn to a permanent its
// controller controls, re-checked as it resolves (CR 608.2b). The second
// sets the colour in layer 5 (CR 613.1e): a spell's through the stack step
// (game.SetSpellColorsForEffect; a permanent spell's permanent keeps it,
// CR 400.7a), a permanent's through an ordinary setColors record. Neither
// needs the source to stay: the effects are the abilities'.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01c8d29f-4924-4d55-80bc-0c0c06bcb733",
		Name:         "Eight-and-a-Half-Tails",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label: "{1}{W}: Target permanent you control gains protection from white until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect,
					Targets: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}).Targets},
				Cost:    effects.ManaCost("{1}{W}"),
				Targets: effects.TargetPermanent("target permanent you control", effects.YouControl()),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"protection from white"},
								Label: "Eight-and-a-Half-Tails — protection from white until end of turn"}.Apply(ctx)
						}
					}
					return nil
				})),
			},
			{
				Label:   "{1}: Target spell or permanent becomes white until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{1}"),
				Targets: effects.TargetSpellOrPermanent("target spell or permanent", nil, nil),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.Gap2B05TargetBecomesColor(ctx, []string{"W"}, "Eight-and-a-Half-Tails — white until end of turn")
				})),
			},
		},
	})
}
