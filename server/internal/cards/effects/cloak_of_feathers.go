package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloak of Feathers — Sorcery {U} (Portal, #5630):
//
//	"Target creature gains flying until end of turn.
//	Draw a card."
//
// Leap's text, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb4baf53-51ed-468b-a468-5d7d45a6dc26",
		Name:         "Cloak of Feathers",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "flying"))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"flying"},
				Label:    "Cloak of Feathers — flying until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
