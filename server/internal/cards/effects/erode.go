package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erode — Instant {W} (EDHREC rank 2092):
//
//	"Destroy target creature or planeswalker. Its controller may
//	 search their library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//
// Path to Exile that destroys instead, and hands over a tapped basic
// as the consolation. The destruction is the single-target verb (so
// indestructible holds, #446), and the search is the S22 chooser
// aimed at the TARGET'S controller — Path to Exile's shape with the
// same "may": they can decline the land, and a controller whose
// library holds no basic searches nothing. The search is offered
// whether or not the destruction stuck, as printed ("Its controller
// may" is not "if it was destroyed"), and it is skipped only when the
// target left before resolution, because then there is no controller
// to ask.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e467fab-e808-44d3-99bf-e3621baeb7cb",
		Name:         "Erode",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:        controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Optional:      true,
				Reason:        "Erode — a basic land card, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
