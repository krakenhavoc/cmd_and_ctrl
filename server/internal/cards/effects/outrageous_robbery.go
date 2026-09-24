package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outrageous Robbery — Instant {X}{B}{B} (Edea steal-and-sac deck,
// #1565):
//
//	"Target opponent exiles the top X cards of their library face down.
//	 You may look at and play those cards for as long as they remain
//	 exiled. If you cast a spell this way, you may spend mana as though
//	 it were mana of any type to cast it."
//
// Gonti, Night Minister's exile-and-play grant with a count of X and a
// target: the top X cards of the target opponent's library, playable
// by the caster (lands included — "play") for as long as each stays in
// exile. X is the announced X (CR 601.2b). An X of zero exiles nothing.
// If the target opponent has left the game by resolution, the spell has
// no legal target and does nothing.
//
// DECLARED SIMPLIFICATIONS, Gonti's two:
//
//   - The cards are exiled FACE UP, so every player sees them. The
//     engine has no face-down exile that only the permission holder may
//     look at.
//   - "Mana of any TYPE" is the engine's "as though it were mana of any
//     colour" permission, which does not cover colourless.
func init() {
	Register(Spec{
		OracleID:     "5b194438-6946-45dd-8d77-c9de8c115d09",
		Name:         "Outrageous Robbery",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The cards are exiled face up, so every player can see them, not only you.",
			"A card whose mana cost includes {C} still needs colorless mana for that part; other mana can pay only its colored and generic parts.",
		},
		XMatters: true,
		Targets:  TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return gontiExileTopForPlay(ctx.Game, t.ID, item.Controller, ctx.X())
			}
			return nil
		},
	})
}
