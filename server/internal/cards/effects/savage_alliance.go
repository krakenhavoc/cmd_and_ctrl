package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Savage Alliance — Instant {2}{R}:
//
//	"Escalate {1} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or more —
//	 • Creatures target player controls gain trample until end of turn.
//	 • Savage Alliance deals 2 damage to target creature.
//	 • Savage Alliance deals 1 damage to each creature target opponent
//	   controls."
//
// Escalate (CR 702.120a, #2126): {1} per mode beyond the first. The
// first and third bullets snapshot the target player's creatures as
// they resolve.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57da1d3f-7469-4efa-9b82-b2d915177d89",
		Name:         "Savage Alliance",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseOneOrMore(
			ModeDoing("Creatures target player controls gain trample until end of turn.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), ControlledBy(t.ID)),
						Keywords: []string{"trample"},
						Label:    "Savage Alliance",
					}.Apply(ctx)
				}),
			ModeDoing("Savage Alliance deals 2 damage to target creature.",
				TargetCreature("target creature"),
				DealFixedDamageToModesTarget(2)),
			ModeWithPurpose(ModeDoing("Savage Alliance deals 1 damage to each creature target opponent controls.",
				TargetPlayer("target opponent", Opponent()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					var ids []game.Card
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == t.ID && c.IsCreature() {
							ids = append(ids, c)
						}
					}
					// One printed instruction is one instance of damage
					// (CR 615.8, ADR 0108 PR 0).
					return ctx.Game.DamageInstanceForEffect(func() error {
						for _, c := range ids {
							if err := (DealDamage{Source: item.SourceCardID, Target: c.InstanceID, Amount: 1}).Apply(ctx.asGroupMember()); err != nil {
								return err
							}
						}
						return nil
					})
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}),
		), EscalateMana("{1}")),
	})
}
