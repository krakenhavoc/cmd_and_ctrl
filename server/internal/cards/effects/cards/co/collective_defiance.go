package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collectiveDefianceRecord is who discarded and how many cards.
const collectiveDefianceRecord = "collective-defiance/discarded"

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
	effects.Register(effects.Spec{
		OracleID:     "43ddc3e8-936b-4be0-9e64-484bbbd69016",
		Name:         "Collective Defiance",
		Completeness: effects.CompletenessFull,
		Modes: effects.Escalating(effects.ChooseOneOrMore(
			// The draw waits for a prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155)
			// and draws the count the discard filed (#4163).
			effects.ModeSteps("Target player discards all the cards in their hand, then draws that many cards.",
				effects.TargetPlayer("target player"),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					n, err := effects.DiscardWholeHand(ctx.Game, t.ID)
					ctx.RecordStep(collectiveDefianceRecord, []uuid.UUID{t.ID}, n)
					return err
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					r := ctx.StepRecord(collectiveDefianceRecord)
					if len(r.IDs) == 0 {
						return nil
					}
					return effects.DrawCards{Player: r.IDs[0], N: r.N()}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Collective Defiance deals 4 damage to target creature.",
				effects.TargetCreature("target creature"),
				effects.DealFixedDamageToModesTarget(4)), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.ModeWithPurpose(effects.ModeDoing("Collective Defiance deals 3 damage to target opponent or planeswalker.",
				effects.TargetOpponentOrPlaneswalker(),
				effects.DealFixedDamageToModesTarget(3)), effects.ForTargets(effects.DamageToTarget(0, 3))),
		), effects.EscalateMana("{1}")),
	})
}
