package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aryel, Knight of Windgrace — Legendary Creature — Human Knight
// {2}{W}{B}, 4/4:
//
//	"Vigilance
//	 {2}{W}, {T}: Create a 2/2 white Knight creature token with
//	 vigilance.
//	 {B}, {T}, Tap X untapped Knights you control: Destroy target
//	 creature with power X or less."
//
// ADR 0109 §9 (#1842): X is the number of Knights the cost taps,
// announced with the activation (CR 602.2b, 107.3a) — the variable tap
// cost the engine already has (TapXUntapped, #1421) — and the power
// bound reads it, before the target is chosen and again as the ability
// resolves (CR 608.2b). Aryel taps for her own {T}, so she is never
// one of the X (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9956f9b7-0140-484c-b606-3685690b84cc",
		Name:            "Aryel, Knight of Windgrace",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{W}, {T}: Create a 2/2 white Knight creature token with vigilance.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1}),
			},
			{
				Label:   "{B}, {T}, Tap X untapped Knights you control: Destroy target creature with power X or less.",
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.TapXUntapped("X untapped Knights you control", effects.OfSubtype("Knight"))),
				Targets: effects.TargetCreature("target creature with power X or less").WithPowerAtMostX(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
