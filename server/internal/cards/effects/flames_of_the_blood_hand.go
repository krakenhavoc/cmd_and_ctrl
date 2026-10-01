package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flames of the Blood Hand — Instant {2}{R}:
//
//	"Flames of the Blood Hand deals 4 damage to target player or
//	 planeswalker. The damage can't be prevented. If that player or that
//	 planeswalker's controller would gain life this turn, that player
//	 gains no life instead."
//
// Two of ADR 0107 §5's pieces. The damage's "can't be prevented" is the
// spell's own rider. The last sentence is a REPLACEMENT ("instead"),
// not "can't gain life": a ModGainNoLife record on the life window for
// the rest of the turn, so another "if you would gain life" replacement
// may be ordered against it (CR 616.1). The player is the target, or
// the planeswalker's controller as the spell resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "4e9df979-c1c2-4de1-944e-c5e2d782e66e",
		Name:                       "Flames of the Blood Hand",
		Completeness:               CompletenessFull,
		SpellDamageCantBePrevented: Always(),
		Targets:                    targetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			player := t.ID
			if t.Kind == game.TargetCard {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				player = c.Controller
			}
			if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.GainNoLifeThisTurnForEffect(item.SourceCardID, player, "Flames of the Blood Hand — gains no life this turn")
			return nil
		},
	})
}
