package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Surprise — Instant {2}{R} (Foundations, #4465):
//
//	"Choose one —
//	 • Creatures you control get +2/+0 until end of turn.
//	 • Create two 1/1 red Goblin creature tokens."
//
// Mode one fixes its set of creatures as the spell resolves (CR 611.2c),
// so a creature that enters later is not pumped (Charge the Sanctum's
// first mode). Mode two makes two of Krenko's Goblins, which count for
// anything that counts Goblins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79cd09e8-d738-47c0-9473-b6281bec6935",
		Name:         "Goblin Surprise",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Creatures you control get +2/+0 until end of turn."),
			ModeWithPurpose(Mode("Create two 1/1 red Goblin creature tokens."), game.Purpose{Tokens: 2, CreatureTokens: 2}),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()),
					Power: 2,
					Label: "Goblin Surprise — +2/+0",
				}.Apply(ctx)
			case ctx.HasMode(1):
				return CreateToken{Controller: item.Controller, Template: RedGoblinToken(), N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
