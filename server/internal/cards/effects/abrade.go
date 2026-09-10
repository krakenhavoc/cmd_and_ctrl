package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abrade — Instant {1}{R}:
//
//	"Choose one —
//	 • Abrade deals 3 damage to target creature.
//	 • Destroy target artifact."
//
// The reason red's two-mana removal slot belongs to Abrade rather
// than a straight burn spell: one card answers either half of the
// board. Both bullets target, which the S20 modal engine allows on a
// "choose one" (Register only rejects two targeted options when Max
// > 1), and the engine applies the CHOSEN option's target clause —
// so the picker offers creatures or artifacts depending on the mode,
// not the union of both.
func init() {
	Register(Spec{
		OracleID: "f9db72dc-9a5b-48a4-a86e-7464d9a2166a",
		Name:     "Abrade",
		Modes: ChooseOne(
			Mode("Abrade deals 3 damage to target creature.",
				TargetCreature("target creature")),
			Mode("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return DealDamage{Source: ctx.Source(), Target: target, Amount: 3}.Apply(ctx)
			case ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
