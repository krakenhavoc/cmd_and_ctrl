package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast Within — Instant {2}{G}:
//
//	"Destroy target permanent. Its controller creates a 3/3 green
//	Beast creature token."
//
// The most-played answer in Commander precisely because the target
// clause is unrestricted: lands, commanders, indestructible-adjacent
// problems, anything. TargetPermanent with no predicate is the whole
// card.
//
// Order matters and is printed order: the permanent is destroyed
// first, then the token is created. That means the controller is
// read BEFORE the destroy — after it, the card is in a graveyard and
// LookupCardForEffect would report a stale controller for a token
// that changed hands. If the target has left the battlefield in
// response, both halves are skipped: no destroy, no token (CR
// 608.2b — the spell has one target, so it fizzles entirely).
func init() {
	Register(Spec{
		OracleID:     "7735eeba-693b-47e2-bd51-414379cf1016",
		Name:         "Beast Within",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
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
			return CreateToken{
				Controller: controller,
				Template:   TokenCard("3/3 green Beast"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
