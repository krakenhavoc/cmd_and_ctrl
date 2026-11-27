package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ricochet Trap — Instant — Trap {3}{R} (Worldwake, S82 OLD b488):
//
//	"If an opponent cast a blue spell this turn, you may pay {R} rather than pay
//	 this spell's mana cost.
//	 Change the target of target spell with a single target."
//
// The trap clause is an alternative cost (CR 118.9), Refraction Trap's shape: offered while
// an opponent's spells cast this turn include a blue one, read off the cast tally as each
// spell was cast (so a countered or resolved blue spell still counts). The change is
// Misdirection's: mandatory, exactly one slot, a target with nowhere else legal to go stays
// as it is (CR 115.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15df3467-c67e-4ae2-9e6f-ffb6ad288a5a",
		Name:         "Ricochet Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell with a single target", effects.HasASingleTarget()),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent cast a blue spell this turn, pay {R}",
			ManaCost:  "{R}",
			Condition: effects.S82OldB488AnOpponentCastABlueSpell,
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.ChangeTargets{
				StackID: item.Targets[0].ID,
				Policy:  game.RetargetChangeOne,
				Reason:  "Ricochet Trap — change the target",
			}.Apply(ctx)
		},
	})
}
