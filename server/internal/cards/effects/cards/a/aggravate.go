package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aggravate — Instant {3}{R}{R} (Avacyn Restored, S82 OLD, #5570):
//
//	"Aggravate deals 1 damage to each creature target player controls. Each
//	 creature dealt damage this way attacks this turn if able."
//
// The creatures are read as the spell resolves and dealt their damage one
// instance each, with Sonic Shrieker's per-recipient continuation: a creature
// that was actually dealt more than 0 (prevention leaves the others alone) and
// is still on the battlefield gains Incite's attack requirement until end of
// turn (CR 508.1d). With its one target illegal the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "192aa7b2-a819-4ffd-bfbf-a270271898e6",
		Name:         "Aggravate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			var creatures []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.Creature()) {
				if c.Controller == t.ID {
					creatures = append(creatures, c.InstanceID)
				}
			}
			return effects.DealDamageToEachThen(ctx, creatures, 1, ctx.Tail(aggravateAttacks))
		},
	})
}

// aggravateAttacks is "each creature dealt damage this way attacks this turn
// if able", told each recipient and what it took. An on-disk key: never
// renamed, never reused.
var aggravateAttacks = effects.RecipientTail("recipient/aggravate-attacks-if-able", func(ctx *effects.Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
	if dealt <= 0 || !effects.OnBattlefield(ctx.Game, target) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Aggravate — attacks this turn if able",
	}.Apply(ctx)
})
