package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assert Perfection — Sorcery {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"Target creature you control gets +1/+0 until end of turn. It deals
//	 damage equal to its power to up to one target creature an opponent
//	 controls."
//
// Burrog Barrage without its condition: two target clauses read by slot
// (Bridgeworks Battle's shape), slot 0 your creature and slot 1 "up to
// one" creature an opponent controls. The +1/+0 always comes first and
// stays when there is no damage to deal; the damage reads your creature's
// layered power after the pump, with that creature as the source, so its
// deathtouch, lifelink or wither apply. If your creature has left, neither
// half happens; a recipient that left takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecf574a5-6410-466d-a726-732134f904b2",
		Name:         "Assert Perfection",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0)),
		},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Assert Perfection — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			// The +1/+0 is a scoped effect the cached view has not seen
			// yet, so the power is read with the layers brought current.
			return DealDamage{Source: mine.ID, Target: theirs.ID, Amount: b31CurrentPowerOf(ctx.Game, mine.ID)}.Apply(ctx)
		},
	})
}
