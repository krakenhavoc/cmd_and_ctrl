package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wesley Crusher, Starfleet Cadet — Legendary Creature — Human Officer
// {2}{U}{U}, 2/2 (Star Trek):
//
//	Flash
//	When Wesley Crusher enters, choose two —
//	• Counter target noncreature spell unless its controller pays {2}.
//	• Return target creature with mana value 2 or less to its owner's
//	  hand.
//	• Target creature can't be blocked this turn.
//	• Target player draws two cards, then discards two cards.
//
// One modal enters trigger asking for exactly two different bullets (CR
// 700.2), each with its own target group, resolved in printed order and
// each target rechecked against its own clause (CR 608.2b). The counter
// is CounterUnlessPaid, asked of the spell's controller. "Can't be blocked
// this turn" is a restriction until end of turn on the one creature. The
// draw-then-discard bullet asks for the discards only after both draws.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Wesley Crusher — choose two",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseN("Choose two", 2, 2,
		effects.ModeDoing("Counter target noncreature spell unless its controller pays {2}.",
			effects.TargetSpell("target noncreature spell", effects.Noncreature()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{2}",
					Question: "Wesley Crusher — pay {2} or your spell is countered",
				}.Apply(ctx)
			}),
		effects.ModeDoing("Return target creature with mana value 2 or less to its owner's hand.",
			effects.TargetCreature("target creature with mana value 2 or less", effects.ManaValueLE(2)),
			effects.BounceTheModesTarget),
		effects.ModeDoing("Target creature can't be blocked this turn.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked,
					Label: "Wesley Crusher — can't be blocked this turn"}.Apply(ctx)
			}),
		effects.ModeWithPurpose(effects.ModeDoing("Target player draws two cards, then discards two cards.",
			effects.TargetPlayer("target player"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return effects.DrawThenDiscard(ctx.Game, t.ID, ctx.Source(), 2, 2, "")
			}),
			effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, Discards: 2})),
	)
	effects.Register(effects.Spec{
		OracleID:        "5c3cbb08-cbe1-42f5-9ac7-e558b85648bb",
		Name:            "Wesley Crusher, Starfleet Cadet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
