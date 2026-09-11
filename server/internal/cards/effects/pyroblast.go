package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyroblast — Instant {R} (EDHREC rank 469):
//
//	"Choose one —
//	 • Counter target spell if it's blue.
//	 • Destroy target permanent if it's blue."
//
// Red Elemental Blast's twin, with the famous difference in the
// wording: "if it's blue" is a condition checked on RESOLUTION, not
// a restriction on the target. So Pyroblast can target any spell or
// permanent — which is what lets it be cast with no blue target in
// sight to grow a storm count or trigger a magecraft — and does
// nothing if the target is not blue when it resolves. REB cannot be
// cast at a non-blue target at all.
//
// The clause is honoured exactly: the target specs are unrestricted
// and OnResolve reads the target's colour before acting.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "ecc435e2-deb1-420a-a79f-01dd08747314",
		Name:     "Pyroblast",
		Modes: ChooseOne(
			Mode("Counter target spell if it's blue.", TargetSpell("target spell")),
			Mode("Destroy target permanent if it's blue.", TargetPermanent("target permanent")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(target)
			if !ok || !card.HasColor("U") {
				return nil
			}
			switch {
			case ctx.HasMode(0):
				return CounterTarget{StackID: target}.Apply(ctx)
			case ctx.HasMode(1):
				if item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
