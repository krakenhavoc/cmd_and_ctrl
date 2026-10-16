package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incite War — Instant {2}{R}:
//
//	"Choose one —
//	 • Creatures target player controls attack this turn if able.
//	 • Creatures you control gain first strike until end of turn.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2} more takes both.
//
// The first bullet is a CR 508.1d attack requirement, Bident of Thassa's
// record scoped to "creatures that player controls" rather than "your
// opponents'" (game.ScopeYourCreatures with the target player as its
// "you", the shape Oriss, Samite Guardian's "can't attack" uses). It
// changes no characteristic and no control, so CR 611.2c does not lock
// its set: a creature that player gains this turn has to attack too. The
// first strike is a fixed set (CR 611.2c), the creatures you control as
// the bullet resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8a640369-59ab-4506-b473-1f804721414c",
		Name:          "Incite War",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Creatures target player controls attack this turn if able.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, t.ID,
						[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, DurationUntilEndOfTurn(ctx),
						"Incite War — creatures that player controls attack this turn if able")
					return nil
				}),
			ModeDoing("Creatures you control gain first strike until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"first strike"},
						Label:    "Incite War — first strike",
					}.Apply(ctx)
				}),
		)),
	})
}
