package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Huddle Up — Sorcery {2}{U}:
//
//	"Assist (Another player can pay up to {2} of this spell's cost.)
//	 Two target players each draw a card."
//
// Axis of Mortality's two-player pick (WithCount(2, 2)), each picked
// player drawing one. A player who has left the game by resolution is
// an illegal target and draws nothing (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:        "81776a16-bf38-450b-995b-beb29121cab2",
		Name:            "Huddle Up",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         TargetPlayer("two target players").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if err := (DrawCards{Player: t.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
