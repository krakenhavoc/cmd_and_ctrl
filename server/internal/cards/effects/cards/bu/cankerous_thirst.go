package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cankerous Thirst — Instant {3}{B/G}:
//
//	"If {B} was spent to cast this spell, you may have target creature get
//	 -3/-3 until end of turn. If {G} was spent to cast this spell, you may have
//	 target creature get +3/+3 until end of turn. (Do both if {B}{G} was
//	 spent.)"
//
// River's Grasp's mana-spent reading for both halves. Each "you may" is a
// resolution yes/no (MayChoice), asked in the printed order, so with {B}{G}
// both on one target the caster can take either or both. The colours are read
// as the spell resolves and carried in the prompt's Word; the target rides in
// Carry. The target is announced with the cast (CR 601.2c), and a target that
// is no longer legal fizzles the whole spell before any question is asked.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "724390aa-240d-4023-9dcb-6bdb2c2090f5",
		Name:           "Cankerous Thirst",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B", "G"},
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			green := ctx.ManaSpentOfColor("G") > 0
			if ctx.ManaSpentOfColor("B") > 0 {
				word := ""
				if green {
					word = "g"
				}
				return effects.S82OldB525Ask(ctx, "Cankerous Thirst — have target creature get -3/-3 until end of turn?", t.ID, cankerousThirstBlack, word)
			}
			if green {
				return effects.S82OldB525Ask(ctx, "Cankerous Thirst — have target creature get +3/+3 until end of turn?", t.ID, cankerousThirstGreen, "")
			}
			return nil
		},
	})
}

var cankerousThirstGreen = effects.S82OldB525MayPump("confirm/cankerous-thirst-green", "Cankerous Thirst — +3/+3 until end of turn", 3, 3, nil)

var cankerousThirstBlack = effects.S82OldB525MayPump("confirm/cankerous-thirst-black", "Cankerous Thirst — -3/-3 until end of turn", -3, -3,
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if r.Word != "g" || len(r.Carry) == 0 {
			return nil
		}
		return effects.S82OldB525Ask(ctx, "Cankerous Thirst — have target creature get +3/+3 until end of turn?", r.Carry[0], cankerousThirstGreen, "")
	})
