package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Long Road Home — Instant {1}{W} (Jumpstart Azorius deck request,
// #4894):
//
//	"Exile target creature. At the beginning of the next end step,
//	 return that card to the battlefield under its owner's control
//	 with a +1/+1 counter on it."
//
// The flicker family's delayed shape (flicker.go): the creature is
// exiled now, and the return is a delayed triggered ability (CR 603.7)
// scheduled from the exile's continuation, over the card that actually
// reached exile. It comes back as a new object (CR 400.7) under its
// OWNER's control, so an opponent's creature you stole goes home, and
// it enters with the +1/+1 counter as part of the entry, so its enters
// triggers see the counter. A token exiled this way does not return
// (CR 111.8); a card that left exile before the end step is not
// followed (CR 603.7c). A target that became illegal in response is
// neither exiled nor scheduled (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8f53018-c496-41f3-bc3f-3c4703b9e4e1",
		Name:         "Long Road Home",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTargetsThenScheduleReturnBody(ctx,
				"Long Road Home — return the exiled card with a +1/+1 counter",
				returnExiledWithAPlusOneCounterBody)
		},
	})
}
