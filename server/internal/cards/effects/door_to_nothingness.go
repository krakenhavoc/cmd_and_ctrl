package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Door to Nothingness — Artifact {5}:
//
//	"This artifact enters tapped.
//	 {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this artifact: Target
//	 player loses the game."
//
// The sacrifice is a cost; the loss is CR 104.3e through LoseTheGame.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b919cca5-88c0-4104-b54e-49644fbf78a5",
		Name:         "Door to Nothingness",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this artifact: Target player loses the game",
			Cost:    Plus(ManaCost("{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}"), TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return LoseTheGame{Player: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
