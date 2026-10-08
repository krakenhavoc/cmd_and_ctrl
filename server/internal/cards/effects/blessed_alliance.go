package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blessed Alliance — Instant {1}{W}:
//
//	"Escalate {2} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or more —
//	 • Target player gains 4 life.
//	 • Untap up to two target creatures.
//	 • Target opponent sacrifices an attacking creature of their
//	   choice."
//
// Escalate (CR 702.120a, #2126): {2} per mode beyond the first. The
// third bullet is an edict restricted to attackers: the opponent
// chooses at resolution, and sacrifices nothing when none of their
// creatures is attacking. The second bullet's "up to two" may name
// none; a target that left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48b67778-6616-4563-8456-62f302649768",
		Name:         "Blessed Alliance",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseOneOrMore(
			ModeWithPurpose(ModeDoing("Target player gains 4 life.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return GainLife{Player: t.ID, Amount: 4}.Apply(ctx)
				}),
				ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 4})),
			ModeDoing("Untap up to two target creatures.",
				TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			ModeDoing("Target opponent sacrifices an attacking creature of their choice.",
				TargetPlayer("target opponent", Opponent()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
						sacrificeSpec("an attacking creature", And(Creature(), attackingCreature())),
						"Blessed Alliance — sacrifice an attacking creature")
					return nil
				}),
		), EscalateMana("{2}")),
	})
}

// attackingCreature passes a permanent that is attacking (CR 506.4):
// unlike AttackingOrBlocking it excludes a blocker.
func attackingCreature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.AttackingTarget != uuid.Nil }
}
