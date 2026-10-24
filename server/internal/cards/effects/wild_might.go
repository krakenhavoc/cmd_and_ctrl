package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Might — Instant {1}{G}:
//
//	"Target creature gets +1/+1 until end of turn. That creature gets
//	 an additional +4/+4 until end of turn unless any player pays {2}."
//
// ADR 0150 (#1903): the +1/+1 lands as the spell resolves; the +4/+4
// waits on the {2} offered to every player in APNAP order from the
// active player (AnyPlayerPays), which holds the step, so it is there
// before damage. A creature that is no longer a legal target gets
// neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f8fa982-73ee-410a-8873-0afc8a6f11fc",
		Name:         "Wild Might",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := wildMightTarget(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: target.ID, Power: 1, Toughness: 1, Label: "Wild Might — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return AnyPlayerPays{
				Cost:     "{2}",
				Question: "Wild Might — pay {2} to stop the additional +4/+4?",
				Then:     wildMightExtra,
			}.Apply(ctx)
		},
	})
}

// wildMightTarget is the spell's creature while it is still a legal
// target.
func wildMightTarget(ctx *Context) (game.TargetRef, bool) {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return t, true
		}
	}
	return game.TargetRef{}, false
}

// wildMightExtra is "if no one does", keyed (#4264): the spell's target rides the asking item.
var wildMightExtra = AnyPlayerConsequences("any-player/wild-might-extra",
	nil,
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		target, ok := wildMightTarget(ctx)
		if !ok {
			return nil
		}
		return BoostUntilEOT{Target: target.ID, Power: 4, Toughness: 4, Label: "Wild Might — an additional +4/+4"}.Apply(ctx)
	},
	nil)
