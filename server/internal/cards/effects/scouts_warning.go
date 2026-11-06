package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scout's Warning — Instant {W} (Avatar: The Last Airbender Eternal, #5075):
//
//	"The next creature card you play this turn can be played as though it
//	 had flash.
//	 Draw a card."
//
// A one-use promise about the next creature spell (game/next_spell_promise.go,
// Savage Summoning's seam) carrying flash alone. No card is chosen as the
// spell resolves; the rule holds until the turn ends or until you cast a
// creature spell, even one cast at a time you could have anyway, which
// spends it (the rulings). Two Scout's Warnings both apply to that one
// next creature spell. A land is played, not cast, and flash never lets a
// land be played at instant speed (the Dryad Arbor ruling), so the promise
// is about creature spells. Then a card is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f65cba6a-3ef4-4841-b2c1-578b2a2c3300",
		Name:         "Scout's Warning",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return GrantNextSpellPromise{From: "Scout's Warning", Promise: game.NextSpellPromise{
					Filter: game.PermissionFilter{CreatureOnly: true},
					Flash:  true,
					Text:   "The next creature card you play this turn can be played as though it had flash.",
				}}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
