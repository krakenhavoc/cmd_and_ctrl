package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farmer Cotton — Legendary Creature — Halfling Peasant {X}{G}{W},
// 1/1 (EDHREC rank 2909):
//
//	"When this creature enters, create X 1/1 white Halfling creature
//	 tokens and X Food tokens. (They're artifacts with "{2}, {T},
//	 Sacrifice this token: You gain 3 life.")"
//
// X Halflings and X Food on a one-drop body. The tokens are made as
// the spell RESOLVES, a beat before Cotton moves from the stack to
// the battlefield — the Goldvein Hydra / Springleaf Parade posture:
// an ETB trigger cannot see the X announced for the spell (the
// stack item is gone by the time the trigger fires), and resolution
// is the last moment X is readable. Halflings first, then Food, as
// printed; the Food is the real Food token.
//
// Sandbox simplification, declared and weaker: the tokens come
// without a trigger on the stack to respond to. Nothing is stronger
// for it — the spell itself could be countered.
func init() {
	Register(Spec{
		OracleID:     "11f1d2b7-0c2b-40df-bf90-3c55b23449af",
		Name:         "Farmer Cotton",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Halflings and Food are created as the spell resolves rather than by an enters trigger you can respond to."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if err := (CreateToken{Controller: item.Controller, Template: b27WhiteHalflingToken(), N: x}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: item.Controller, Template: FoodToken(), N: x}.Apply(ctx)
		},
	})
}
