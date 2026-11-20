package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voices from the Void — Sorcery {4}{B} (Conflux, #5587):
//
//	"Domain — Target player discards a card for each basic land type
//	 among lands you control."
//
// The count is the shared DomainCount (domain.go, CR 305.6), read as the
// spell resolves. The discarding player picks the cards (CR 701.9a), so
// it opens a real prompt; with no basic land type nothing is discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c5264d2-8544-4e83-8ebc-7261dde80f01",
		Name:         "Voices from the Void",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			p, ok := firstLegalPlayerTarget(ctx)
			if n := DomainCount(ctx.Game, item.Controller); ok && n > 0 {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p, Source: item.SourceCardID, N: n})
			}
			return nil
		},
	})
}
