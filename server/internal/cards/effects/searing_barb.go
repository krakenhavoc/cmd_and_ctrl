package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			t := legal[0]
			isCreature := false
			if t.Kind == game.TargetCard {
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.IsCreature() {
					isCreature = true
				}
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			if isCreature {
				if z := ctx.Game.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneBattlefield {
					if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock,
						Label: "Searing Barb — can't block"}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return Incubate{N: 1}.Apply(ctx)
		},
	})
}
