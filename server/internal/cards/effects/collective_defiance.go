package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collective Defiance — Sorcery {1}{R}{R}:
//
//	"Escalate {1} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or more —
//	 • Target player discards all the cards in their hand, then draws
//	   that many cards.
//	 • Collective Defiance deals 4 damage to target creature.
//	 • Collective Defiance deals 3 damage to target opponent or
//	   planeswalker."
//
// Escalate (CR 702.120a, #2126) is {1} per mode beyond the first. The
// first bullet discards the whole hand (every discard before any draw,
// so discard payoffs queue first) and draws as many as were discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43ddc3e8-936b-4be0-9e64-484bbbd69016",
		Name:         "Collective Defiance",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseOneOrMore(
			ModeDoing("Target player discards all the cards in their hand, then draws that many cards.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					n, err := discardWholeHand(ctx.Game, t.ID)
					if err != nil {
						return err
					}
					return DrawCards{Player: t.ID, N: n}.Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Collective Defiance deals 4 damage to target creature.",
				TargetCreature("target creature"),
				DealFixedDamageToModesTarget(4)), ForTargets(DamageToTarget(0, 4))),
			ModeWithPurpose(ModeDoing("Collective Defiance deals 3 damage to target opponent or planeswalker.",
				targetOpponentOrPlaneswalker(),
				DealFixedDamageToModesTarget(3)), ForTargets(DamageToTarget(0, 3))),
		), EscalateMana("{1}")),
	})
}
