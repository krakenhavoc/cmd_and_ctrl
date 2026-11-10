package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coordinated Maneuver — Instant {1}{W} (Tarkir: Dragonstorm):
//
//	"Choose one —
//	 • Coordinated Maneuver deals damage equal to the number of creatures
//	   you control to target creature or planeswalker.
//	 • Destroy target enchantment."
//
// The first mode counts the creatures its controller controls as the
// spell resolves (CR 608.2h), the target creature included; the second is
// a plain destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ff39311-565e-4414-b32b-eb24c70b3090",
		Name:         "Coordinated Maneuver",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			Mode("Coordinated Maneuver deals damage equal to the number of creatures you control to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
			Mode("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				n := countControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsCreature() })
				return DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
			case ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
