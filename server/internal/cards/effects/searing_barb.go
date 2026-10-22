package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Searing Barb — Sorcery {2}{R} (March of the Machine):
//
//	"Searing Barb deals 2 damage to any target. If it's a creature, it
//	 can't block this turn. Incubate 1."
//
// "If it's a creature" is read before the damage, while the target is
// still on the battlefield, and the restriction lands whether or not
// the damage killed or was prevented, as printed (a dead creature has
// nothing to restrict, and the restriction is skipped for it). The
// incubate (CR 701.53a, ADR 0185) happens if the spell resolves at all,
// which requires its target to be legal still.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a1622a5-c982-407b-a158-7b8ad9238aac",
		Name:         "Searing Barb",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		// The restriction and the Incubator wait for a prompt the damage asks (a CR 616
		// order; CR 608.2c, #4155). Whether the target was a creature is read before the
		// damage and filed for the steps behind it.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				t := legal[0]
				isCreature := 0
				if t.Kind == game.TargetCard {
					if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.IsCreature() {
						isCreature = 1
					}
				}
				ctx.RecordStep(searingBarbRecord, []uuid.UUID{t.ID}, isCreature)
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				r := ctx.StepRecord(searingBarbRecord)
				if len(r.IDs) == 0 || len(r.Ns) == 0 || r.Ns[0] == 0 {
					return nil
				}
				if z := ctx.Game.FindCardZoneForEffect(r.IDs[0]); z != nil && z.Kind == game.ZoneBattlefield {
					return RestrictUntilEOT{Target: r.IDs[0], Restrictions: game.CantBlock,
						Label: "Searing Barb — can't block"}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if len(ctx.StepRecord(searingBarbRecord).IDs) == 0 {
					return nil
				}
				return Incubate{N: 1}.Apply(ctx)
			}),
		),
	})
}

// searingBarbRecord files the target, and 1 when it was a creature, before the damage.
const searingBarbRecord = "searing-barb/target"
