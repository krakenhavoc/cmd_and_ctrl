package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Outmaneuver — Instant {X}{R} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"X target blocked creatures assign their combat damage this turn as
//	 though they weren't blocked."
//
// Exactly X target blocked creatures (CountFromX; BlockedCreature is CR 509.1h,
// so a creature whose blockers have all left combat still counts). Each one
// still legal as the spell resolves gets, for the turn, the mandatory
// "assigns its combat damage as though it weren't blocked" (CR 510.1c,
// game.MustAssignAsThoughUnblockedMod): unlike Thorn Elemental's "you may" it
// is not a choice, so all of the creature's damage goes to what it attacks
// and none to its blockers. The grant goes on the creatures chosen, not on
// whatever is blocked later (CR 611.2c).
//
// No simplification.
func init() {
	spec := effects.TargetCreature("X target blocked creatures", effects.BlockedCreature())
	spec.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "42b44108-4a03-4333-bf5a-e42493f4d5e4",
		Name:         "Outmaneuver",
		Completeness: effects.CompletenessFull,
		Targets:      spec,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.MustAssignAsThoughUnblockedMod()},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Outmaneuver — assigns combat damage as though it weren't blocked",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
