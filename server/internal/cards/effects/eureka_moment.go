package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eureka Moment — Instant {2}{G}{U} (EDHREC rank 1986):
//
//	"Draw two cards. You may put a land card from your hand onto the
//	 battlefield."
//
// Instant-speed card draw with a free land drop attached. The draw
// is the whole of what runs.
//
// DECLARED SIMPLIFICATION — the Spelunking / Insidious Fungus
// posture: "you may put a land card from your hand onto the
// battlefield" is not implemented. It is a pick-from-hand prompt
// with a hand-to-battlefield move, and neither exists (the Growth
// Spiral gap; the only hand picker the engine has is the discard
// modal). Shipping it as "put the first land" would be a choice the
// player never made, so the spell draws its two cards and stops —
// weaker than printed, never stronger. It becomes whole the day a
// put-from-hand prompt lands.
func init() {
	Register(Spec{
		OracleID:     "0e2c11b2-d95f-4402-9a4a-afd3f7ffb8be",
		Name:         "Eureka Moment",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only the two cards are drawn — it doesn't offer to put a land from your hand onto the battlefield."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
