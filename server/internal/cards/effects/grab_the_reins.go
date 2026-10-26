package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grab the Reins — Instant {3}{R}:
//
//	"Choose one —
//	 • Until end of turn, you gain control of target creature and it
//	   gains haste.
//	 • Sacrifice a creature. Grab the Reins deals damage equal to that
//	   creature's power to any target.
//	 Entwine {2}{R} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2}{R} more takes both, and the bullets run
// in the order written (CR 702.42b) — the classic play: steal a
// creature, then sacrifice it.
//
// The first bullet is Act of Treason without the untap: a layer-2 theft
// until end of turn plus a haste grant. The second bullet's sacrifice is
// not a cost: it is chosen as the spell resolves, among the creatures
// you control then (the Disciple of Bolas shape), and "that creature's
// power" is its last-known power (CR 608.2h), floored at zero. Its
// "any target" was chosen on the cast; if that target is illegal by
// resolution the creature is still sacrificed and no damage is dealt
// (CR 608.2b — the rest of the spell still does what it can).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3ceda405-8175-416a-beda-9a95f5acc60c",
		Name:          "Grab the Reins",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}{R}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Until end of turn, you gain control of target creature and it gains haste.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					if err := (GainControl{
						Target:   t.ID,
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Grab the Reins — gain control until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Grab the Reins — haste"}.Apply(ctx)
				}),
			ModeDoing("Sacrifice a creature. Grab the Reins deals damage equal to that creature's power to any target.",
				TargetAny(), grabTheReinsFling),
		)),
	})
}

// grabTheReinsFling is the second bullet: the caster sacrifices a
// creature they control, then the bullet's own target — if still legal —
// takes damage equal to its last-known power.
func grabTheReinsFling(item *game.StackItem, ctx *Context, occ int) error {
	target, legal := ModeTarget(ctx, occ)
	// Entwined, the first bullet's theft is a fresh layer-2 record
	// in this same resolution: settle the layers so the creature
	// it took is among "a creature" you control to sacrifice.
	ctx.Game.RecomputeLayersIfStaleLocked()
	// Keyed (#4290): the bullet's target, when still legal, rides the run
	// as its carry.
	var carry []uuid.UUID
	if legal {
		carry = []uuid.UUID{target.ID}
	}
	return ctx.SacrificeThen(ctx.Controller(),
		sacrificeSpec("a creature", Creature()),
		"Grab the Reins — sacrifice a creature", 1, grabTheReinsDamage, carry...)
}

// grabTheReinsDamage is "Grab the Reins deals damage equal to that
// creature's power to any target" once the creature has really been
// sacrificed, keyed (#4290): the target is the run's carry.
var grabTheReinsDamage = RunThen("prompt-run/grab-the-reins-damage", func(ctx *Context, r game.RunDone) error {
	ids := r.Sacrifices().By(ctx.Controller())
	if len(ids) == 0 || len(r.Carry) == 0 {
		return nil
	}
	power := departedCreaturePower(ctx.Game, ids[0])
	if power <= 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: power}.Apply(ctx)
})
