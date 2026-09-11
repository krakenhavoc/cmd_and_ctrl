package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampiric Tutor — "Search your library for a card, then shuffle
// and put that card on top of your library. You lose 2 life."
//
// S14 sandbox simplifications:
//   - The tutored card goes directly to HAND rather than on top of
//     the library. The "put on top + draw next turn" nuance matters
//     when racing or dodging discard; the net-card-selection effect
//     is identical. Revisit in S17 when we model replacement timing
//     / mulligan-in-game flows.
//
// The controller picks which card (S22).
func init() {
	Register(Spec{
		OracleID:     "ededbdae-d9dc-4206-9335-d7158f2d7700",
		Name:         "Vampiric Tutor",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The card you find goes straight to your hand instead of on top of your library."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  false,
				Shuffle: true,
				Reason:  "Vampiric Tutor — search your library for a card",
			}).Apply(ctx); err != nil {
				return err
			}
			// The life loss is a separate sentence and does not wait
			// on the search, so it stays inline rather than riding
			// the Then continuation.
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
