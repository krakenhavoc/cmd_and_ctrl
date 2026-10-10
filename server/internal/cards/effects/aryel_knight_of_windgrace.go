package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "9956f9b7-0140-484c-b606-3685690b84cc",
		Name:            "Aryel, Knight of Windgrace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{W}, {T}: Create a 2/2 white Knight creature token with vigilance.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
				Effect:  Do(CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1}),
			},
			{
				Label:   "{B}, {T}, Tap X untapped Knights you control: Destroy target creature with power X or less.",
				Cost:    Plus(ManaCost("{B}"), TapCost(), TapXUntapped("X untapped Knights you control", OfSubtype("Knight"))),
				Targets: TargetCreature("target creature with power X or less").WithPowerAtMostX(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return DestroyTarget{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
