package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flameblast Dragon — Creature — Dragon {4}{R}{R}, 5/5:
//
//	"Flying
//	 Whenever this creature attacks, you may pay {X}{R}. If you do, it
//	 deals X damage to any target."
//
// The target is chosen as the trigger is put on the stack (CR 603.3d),
// and X as it resolves (CR 608.2d), through MayPayX (#2727): a number
// from 0 to the most you can pay, then the ordinary "you may pay {3}{R}"
// prompt. A bot pays for the target's lethal damage when it can (an
// opponent's creature or planeswalker), everything it can at an
// opponent, and nothing at itself or its own permanents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3180ee4f-66d7-4b73-bc10-9ee0c790ec18",
		Name:            "Flameblast Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Flameblast Dragon — pay {X}{R} to deal X damage to any target", flameblastDragonAttacks),
				TargetAny()),
		},
	})
}

func flameblastDragonAttacks(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}{R}",
		Label: "Flameblast Dragon",
		Buys:  "deal X damage to the target",
		Goal:  flameblastDragonGoal,
		Unit:  game.PayAmountDamage,
		OnPay: func(ctx *Context, x int) error {
			return sourceDealsComputedDamageToEachLegalTarget(func(*Context) int { return x })(ctx.Game, ctx.Item)
		},
	}.Apply(NewContext(g, item))
}

// flameblastDragonGoal is the X a bot pays: the target's lethal damage
// when it is an opponent's permanent and that much can be paid, every
// point at an opponent, and nothing at you or your own permanents.
func flameblastDragonGoal(ctx *Context, ceiling int) int {
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return 0
	}
	you := ctx.Controller()
	id := targets[0].ID
	if p := ctx.Game.PlayerByIDForEffect(id); p != nil {
		if p.ID == you {
			return 0
		}
		return ceiling
	}
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.Controller == you {
		return 0
	}
	if lethal := lethalDamageGoal(ctx); lethal > 0 && lethal <= ceiling {
		return lethal
	}
	return 0
}
