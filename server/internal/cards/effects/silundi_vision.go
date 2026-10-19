package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silundi Vision // Silundi Isle — Modal DFC. Front: Instant {2}{U}:
// "Look at the top six cards of your library. You may reveal an instant
// or sorcery card from among them and put it into your hand. Put the rest
// on the bottom of your library in a random order." Back: Land
// (mdfc_lands.go).
//
// The look-and-take sentence helper (Horn of the Mark's shape): only the
// caster sees the six, the card taken is revealed, and the rest go to the
// bottom in a random order once the answer is in. No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0182ca0-f353-4012-9121-6f4ac9f7a046",
		Name:         "Silundi Vision",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return LookAtTopThenMayTakeToHand(6, Or(Instant(), Sorcery()), 1,
				"Silundi Vision — you may reveal an instant or sorcery card from among them and put it into your hand")(ctx.Game, item)
		},
	})
}
