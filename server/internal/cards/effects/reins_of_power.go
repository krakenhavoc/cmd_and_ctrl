package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reins of Power — Instant {2}{U}{U}:
//
//	"Untap all creatures you control and all creatures target
//	 opponent controls. You and that opponent each gain control of
//	 all creatures the other controls until end of turn. Those
//	 creatures gain haste until end of turn."
//
// A two-way Insurrection: both snapshots (the caster's creatures and
// the chosen opponent's) are taken from who controls what BEFORE
// anything moves, so the untap and the swap both read the original
// board rather than one another's output. The swap itself is two runs
// of GainControl — the caster's creatures go to the opponent, the
// opponent's go to the caster — because ExchangeControl (control.go)
// is CR 701.12's ALL-OR-NOTHING single-pair primitive and this clause
// exchanges two disjoint SETS, whatever size each is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96381388-1e82-411c-8290-1f3e909d3b5f",
		Name:         "Reins of Power",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			caster := ctx.Controller()
			mine := MatchingBattlefield(ctx, And(Creature(), YouControl()))
			theirs := MatchingBattlefield(ctx, reinsOfPowerControlledBy(victim))
			for _, c := range mine {
				if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, c := range theirs {
				if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := reinsOfPowerSwap(ctx, mine, victim); err != nil {
				return err
			}
			return reinsOfPowerSwap(ctx, theirs, caster)
		},
	})
}

// reinsOfPowerControlledBy is a MatchingBattlefield predicate for "all
// creatures <player> controls" — the ignore-caster shape a target
// player's own creatures need, since MatchingBattlefield always hands
// the resolving spell's caster to the predicate and this half of the
// clause cares about someone else.
func reinsOfPowerControlledBy(player uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.Controller == player
	}
}

// reinsOfPowerSwap hands one snapshot of creatures to `to`: gain
// control until end of turn, then haste until end of turn, for each.
func reinsOfPowerSwap(ctx *Context, creatures []game.Card, to uuid.UUID) error {
	for _, c := range creatures {
		if err := (GainControl{
			Target:     c.InstanceID,
			Controller: to,
			Duration:   DurationUntilEndOfTurn(ctx),
			Label:      "Reins of Power — gain control until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, c := range creatures {
		if err := (GrantKeywordUntilEOT{
			Target:   c.InstanceID,
			Keywords: []string{"haste"},
			Label:    "Reins of Power — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
