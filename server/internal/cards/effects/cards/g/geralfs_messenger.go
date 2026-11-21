package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geralf's Messenger — Creature — Zombie {B}{B}{B}, 3/2:
//
//	"This creature enters tapped.
//	 When this creature enters, target opponent loses 2 life.
//	 Undying"
//
// It enters tapped every time, including when undying returns it, and
// each entry drains the chosen opponent for 2. Undying is
// PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "740f9740-aaa7-4061-86f5-85be79742541",
		Name:            "Geralf's Messenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Geralf's Messenger — target opponent loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.GainLife{Player: t.ID, Amount: -2}.Apply(ctx)
					}
					return nil
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
