package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparkhunter Masticore — Artifact Creature — Masticore {3}, 3/4 (Core Set
// 2021, #5511):
//
//	"As an additional cost to cast this spell, discard a card.
//	 Protection from planeswalkers
//	 {1}: This creature deals 1 damage to target planeswalker.
//	 {3}: This creature gains indestructible until end of turn."
//
// The discard is an additional cost paid on the cast. Protection from
// planeswalkers is the protection grammar's card-type quality, so a
// planeswalker can't damage, enchant, block or target it; it can still
// damage a planeswalker itself. The ping is repeatable and targets a
// planeswalker only; the indestructible grant lasts until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1306c72f-fc6b-470b-a95e-0364e0e30ac7",
		Name:            "Sparkhunter Masticore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from planeswalkers"},
		AdditionalCost:  effects.DiscardCost(1),
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}: This creature deals 1 damage to target planeswalker.",
				Cost:    effects.ManaCost("{1}"),
				Targets: effects.TargetPermanent("target planeswalker", effects.Planeswalker()),
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
				})),
			},
			{
				Label:   "{3}: This creature gains indestructible until end of turn.",
				Cost:    effects.ManaCost("{3}"),
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Sparkhunter Masticore — indestructible until end of turn"),
			},
		},
	})
}
