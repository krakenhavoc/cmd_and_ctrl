package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consider the Prime Directive — Instant {2}{U} (Star Trek, #4973):
//
//	"Choose one —
//	 • Interfere — Counter target spell unless its controller pays {3}.
//	 • Observe — Tap up to one target creature and put a stun counter on
//	   it. Draw a card. (If a permanent with a stun counter would become
//	   untapped, remove one from it instead.)"
//
// "Interfere" and "Observe" are mode names with no rules meaning. The
// first bullet is Decisive Denial's CounterUnlessPaid over any spell; the
// second taps and stuns its target (none is a legal choice) and draws
// either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f1d416c-60de-4796-a513-097d1d9a2305",
		Name:         "Consider the Prime Directive",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Interfere — Counter target spell unless its controller pays {3}.",
				TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{3}",
						Question: "Consider the Prime Directive — pay {3} or your spell is countered",
					}.Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Observe — Tap up to one target creature and put a stun counter on it. Draw a card.",
				UpToOneTargetCreature("up to one target creature"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					if t, ok := ModeTarget(ctx, occ); ok {
						if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}), game.Purpose{Draws: 1}),
		),
	})
}
