package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silence — Instant {W} (EDHREC top 1,000, #3654):
//
//	"Your opponents can't cast spells this turn."
//
// Mandate of Peace's second sentence on its own: an outright cast ban
// (game.CastBanOutright) granted to each opponent until end of turn
// (CR 611.2). A spell an opponent already has on the stack is
// unaffected; they just cannot cast another.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8aed54cb-d1bb-45ad-adbe-38e55d84ff31",
		Name:         "Silence",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			d := DurationUntilEndOfTurn(ctx)
			for _, opp := range ctx.Opponents() {
				if err := (RestrictCasting{
					Player:   opp,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright},
					Label:    "Silence — can't cast spells this turn",
					Duration: d,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
