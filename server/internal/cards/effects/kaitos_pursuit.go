package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaito's Pursuit — Sorcery {2}{B} (Kamigawa: Neon Dynasty, #5493):
//
//	"Target player discards two cards. Ninjas and Rogues you control gain
//	 menace until end of turn."
//
// The target chooses what to discard (CR 701.9a) and the table waits for
// it. The menace goes to the Ninjas and Rogues the caster controls as the
// spell resolves (CR 611.2c), whether or not the discard did anything.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23b7fd18-3129-448f-881f-947ba988f98d",
		Name:         "Kaito's Pursuit",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: item.Targets[0].ID,
				Source: item.SourceCardID,
				N:      2,
			})
			you := ctx.Controller()
			return GrantKeywordUntilEOT{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.Controller == you && c.IsCreature() && (c.HasSubtypeWord("Ninja") || c.HasSubtypeWord("Rogue"))
				},
				Keywords: []string{"menace"},
				Label:    "Kaito's Pursuit — menace",
			}.Apply(ctx)
		},
	})
}
