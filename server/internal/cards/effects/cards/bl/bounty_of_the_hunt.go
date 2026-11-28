package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounty of the Hunt — Instant {3}{G}{G} (Alliances, S82 OLD batch 817):
//
//	"You may exile a green card from your hand rather than pay this
//	 spell's mana cost.
//	 Distribute three +1/+1 counters among one, two, or three target
//	 creatures. For each +1/+1 counter you put on a creature this way,
//	 remove a +1/+1 counter from that creature at the beginning of the next
//	 cleanup step."
//
// The pitch is Misdirection's with a green card. One delayed trigger per
// creature, each carrying how many counters it was given; a target that left
// the battlefield in response was given none and gets no trigger.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77a43c24-398f-42c6-a0d3-b28b33a0ce12",
		Name:         "Bounty of the Hunt",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Pitch("Exile a green card from your hand", 0,
				effects.CardInYourHand("a green card from your hand", effects.OfColor("G")),
				"a green card from your hand"),
		},
		Targets: effects.TargetCreature("one, two, or three target creatures").WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			shares := item.Distribution
			for _, t := range ctx.LegalTargets() {
				n := shares[t.ID]
				if t.Kind != game.TargetCard || n <= 0 {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
					return err
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				if err := (effects.ScheduleDelayedTrigger{
					At:     game.StepCleanup,
					Label:  "Bounty of the Hunt — remove the counters from " + c.Name,
					Cards:  []uuid.UUID{t.ID},
					Body:   effects.OldB817RemoveCountersAtCleanupBody,
					Params: game.EffectParams{Object: game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch}, Amount: n},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
