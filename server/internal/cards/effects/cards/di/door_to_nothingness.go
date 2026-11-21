package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b919cca5-88c0-4104-b54e-49644fbf78a5",
		Name:         "Door to Nothingness",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this artifact: Target player loses the game",
			Cost:    effects.Plus(effects.ManaCost("{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.LoseTheGame{Player: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
