package effects

// Hymn of the Faller — Sorcery {1}{B} (Edge of Eternities):
//
//	"Surveil 1, then you draw a card and lose 1 life. (To surveil 1, look
//	 at the top card of your library. You may put it into your
//	 graveyard.)
//	 Void — If a nonland permanent left the battlefield this turn or a
//	 spell was warped this turn, draw another card."
//
// Four steps in the printed order, each waiting for a prompt the one
// before it asks (CR 608.2c): the surveil decision is made before the draw,
// so the card it leaves on top is the card drawn. Void (void.go) is read at
// its own step, after the life loss, as the spell resolves; nothing the
// earlier steps do can put a permanent in or out of play.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab1379b7-0016-4d2a-a83b-952666e43b89",
		Name:         "Hymn of the Faller",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Surveil{N: 1},
			DrawCards{N: 1},
			youLoseLife(1),
			Instruction(func(ctx *Context) error {
				if !Void(ctx.Game) {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
