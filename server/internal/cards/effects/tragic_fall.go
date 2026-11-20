package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tragic Fall — Instant {1}{B} (Modern Horizons 2, #5501):
//
//	"Target creature gets -3/-3 until end of turn.
//	 Hellbent — That creature gets -13/-13 until end of turn instead if
//	 you have no cards in hand."
//
// Last Gasp's -3/-3, with Demonfire's hellbent condition (SpellHellbent)
// read as the spell resolves, when Tragic Fall itself has already left
// the hand.
//
// No simplification.
func init() {
	hellbent := SpellHellbent()
	Register(Spec{
		OracleID:     "1b7e325d-24ee-42e3-b06f-6c91ec5280c3",
		Name:         "Tragic Fall",
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			n := 3
			if hellbent(ctx.Game, item) {
				n = 13
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -n,
				Toughness: -n,
				Label:     "Tragic Fall — -N/-N until end of turn",
			}.Apply(ctx)
		},
	})
}
