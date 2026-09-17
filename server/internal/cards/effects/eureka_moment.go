package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eureka Moment — Instant {2}{G}{U} (EDHREC rank 1986):
//
//	"Draw two cards. You may put a land card from your hand onto the
//	 battlefield."
//
// Instant-speed card draw with a free land drop attached.
//
// Both halves run, in printed order: the draws first, so a land just
// drawn is a legal pick, then the offer. The land drop is the shared
// MayPutALandFromHand clause (#654) — a pick-from-hand prompt whose
// continuation puts the land onto the battlefield through the CR 614
// entry pipeline. It is not a land PLAY (CR 305.4), so the turn's
// land drop is still available afterwards.
//
// Until #654 the clause was omitted and declared: the prompt existed
// (#552) but the hand-to-battlefield move did not, and shipping it as
// "put the first land" would have been a choice the player never
// made.
func init() {
	Register(Spec{
		OracleID:     "0e2c11b2-d95f-4402-9a4a-afd3f7ffb8be",
		Name:         "Eureka Moment",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return MayPutALandFromHand("Eureka Moment").Apply(ctx)
		},
	})
}
