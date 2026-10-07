package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rush of Inspiration // Crackling Falls — modal double-faced card. This
// file is the FRONT face, Instant {1}{U/R}{U/R}:
//
//	"Draw two cards. Then discard a card at random unless you pay
//	 {E}{E} (two energy counters)."
//
// The back face, Crackling Falls, is registered with the MDFC land cycle
// in mdfc_lands.go.
//
// ADR 0129 §3 (#1995): the two cards are drawn, then the energy is asked
// for through the pay-unless prompt (CR 118.12a); declining, or being
// short (CR 118.3), discards one card at random.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbd569cc-bc21-46df-b8eb-5b5bcd8fe762",
		Name:         "Rush of Inspiration",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 2}).Apply(ctx); err != nil {
				return err
			}
			return PayEnergyUnless{
				N:        2,
				Question: "Rush of Inspiration — pay {E}{E}, or discard a card at random?",
				OnDecline: func(ctx *Context) error {
					return DiscardCards{N: 1}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
