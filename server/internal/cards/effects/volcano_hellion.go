package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volcano Hellion — Creature — Hellion {2}{R}{R}, 6/5:
//
//	"This creature has echo {X}, where X is your life total. (At the
//	 beginning of your upkeep, if this came under your control since the
//	 beginning of your last upkeep, sacrifice it unless you pay its echo
//	 cost.)
//	 When this creature enters, it deals an amount of damage of your
//	 choice to you and target creature. The damage can't be prevented."
//
// The echo is EchoX (ADR 0108 §5), read as the trigger resolves. The
// amount is chosen as the enters trigger resolves (CR 608.2d), on the
// number prompt of ADR 0129's amendment of 2026-10-09 (#1941): any
// number from 0 up, with no ceiling, so an amount past lethal is there
// for a lifelink or "whenever this is dealt damage" target. The target
// is chosen as the trigger goes on the stack, and the damage to you and
// to it is one damage instruction (CR 615.8) that can't be prevented
// (CR 615.12, ADR 0107 §5).
//
// A bot is offered 0, the target's lethal damage and its own life total,
// and deals the lethal amount only to an opponent's creature and only
// while that leaves it at 10 life or more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12a88c08-11e3-4f51-a17c-ca06452bac6f",
		Name:         "Volcano Hellion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			EchoX("Volcano Hellion", "your life total", func(ctx *Context) int {
				if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil {
					return p.Life
				}
				return 0
			}),
			Targeting(WhenThisEnters("Volcano Hellion — damage of your choice to you and target creature", volcanoHellionDamage),
				TargetCreature("target creature")),
		},
	})
}

// volcanoHellionDamage asks for the amount, then deals it to the
// trigger's controller and to the target.
func volcanoHellionDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	target := targets[0].ID
	you := ctx.Controller()
	source := ctx.Source()
	var sourceObj *game.ObjectRef
	if ref, ok := ctx.SourceRef(); ok {
		sourceObj = &ref
	}
	return ChooseNumber{
		Question:   "Volcano Hellion — choose an amount of damage to deal to you and the target creature",
		NoMax:      true,
		Unit:       game.PayAmountDamage,
		SelfDamage: true,
		Goal: func(ctx *Context) int {
			if c, ok := ctx.Game.LookupCardForEffect(target); !ok || c.Controller == you {
				return 0
			}
			return lethalDamageGoal(ctx)
		},
		Marks: func(ctx *Context) []int {
			marks := []int{lethalDamageGoal(ctx)}
			if p := ctx.Game.PlayerByIDForEffect(you); p != nil {
				marks = append(marks, p.Life)
			}
			return marks
		},
		Then: func(ctx *Context, n int) error {
			return volcanoHellionDeal(ctx, source, sourceObj, you, target, n)
		},
	}.Apply(ctx)
}

// volcanoHellionDeal is "it deals that much damage to you and target
// creature. The damage can't be prevented": one damage instruction, two
// recipients.
func volcanoHellionDeal(ctx *Context, source uuid.UUID, obj *game.ObjectRef, you, target uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, to := range []uuid.UUID{you, target} {
			if err := (DealDamage{Source: source, SourceObject: obj, Target: to, Amount: n, CantBePrevented: true}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
