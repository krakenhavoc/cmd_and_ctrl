package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Super-Skrull — {1}{B}{B}{B} Legendary Creature — Skrull Shapeshifter
// Villain 4/5 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 {2}{W}: Create a 0/4 colorless Wall creature token with defender.
//	 {3}{G}: Super-Skrull gets +4/+4 until end of turn.
//	 {4}{R}: Super-Skrull deals 4 damage to target creature.
//	 {5}{U}: Target player draws four cards."
//
// Four independent activated rows, each an existing shape: a token maker
// (the Wall is the Fantastic Four's), a self pump, targeted damage and a
// targeted draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4644013f-c96a-450e-a7bc-ddfa7a563050",
		Name:            "Super-Skrull",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{W}: Create a 0/4 colorless Wall creature token with defender.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
				Cost:    effects.ManaCost("{2}{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("0/4 colorless Wall with defender"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}{G}: Super-Skrull gets +4/+4 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 4, Toughness: 4}},
				Cost:    effects.ManaCost("{3}{G}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(4, 4, "Super-Skrull — +4/+4 until end of turn"),
			},
			{
				Label:   "{4}{R}: Super-Skrull deals 4 damage to target creature.",
				Cost:    effects.ManaCost("{4}{R}"),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 4}.Apply(ctx)
				},
			},
			{
				Label:   "{5}{U}: Target player draws four cards.",
				Cost:    effects.ManaCost("{5}{U}"),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.DrawCards{Player: t.ID, N: 4}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
