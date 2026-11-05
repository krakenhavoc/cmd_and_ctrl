package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trigger := WhenThisEnters("Wesley Crusher — choose two",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseN("Choose two", 2, 2,
		ModeDoing("Counter target noncreature spell unless its controller pays {2}.",
			TargetSpell("target noncreature spell", Noncreature()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return CounterUnlessPaid{
					StackID:  t.ID,
					Cost:     "{2}",
					Question: "Wesley Crusher — pay {2} or your spell is countered",
				}.Apply(ctx)
			}),
		ModeDoing("Return target creature with mana value 2 or less to its owner's hand.",
			TargetCreature("target creature with mana value 2 or less", ManaValueLE(2)),
			BounceTheModesTarget),
		ModeDoing("Target creature can't be blocked this turn.",
			TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked,
					Label: "Wesley Crusher — can't be blocked this turn"}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("Target player draws two cards, then discards two cards.",
			TargetPlayer("target player"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return drawThenDiscard(ctx.Game, t.ID, ctx.Source(), 2, 2, "")
			}),
			ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, Discards: 2})),
	)
	Register(Spec{
		OracleID:        "5c3cbb08-cbe1-42f5-9ac7-e558b85648bb",
		Name:            "Wesley Crusher, Starfleet Cadet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
