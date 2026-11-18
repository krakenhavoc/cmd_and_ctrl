package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storyweave — Instant {2}{G} (Kamigawa: Neon Dynasty, S82 #5493):
//
//	"Choose one —
//	 • Put two +1/+1 counters on target creature you control.
//	 • Put two lore counters on target Saga you control. The next time
//	   one or more enchantment creatures you control enter this turn,
//	   each enters with two additional +1/+1 counters on it."
//
// The lore counters go on first (a chapter ability triggers as they are
// put on, CR 714.2b), then the rest of the bullet: a one-shot grant on
// the controller for the turn. It is spent by the first entry of an
// enchantment creature under their control, and by the other enchantment
// creatures that enter together with it, and by nothing else (a
// nonenchantment creature entering first does not use it up). The counters
// are put on as the creature enters (CR 614.1c), through the counter
// pipeline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94389047-4976-4be2-aaca-25dcdc538abe",
		Name:         "Storyweave",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Put two +1/+1 counters on target creature you control.",
				TargetCreature("target creature you control", YouControl()),
				twoPlusOneCountersOnModeTarget),
			ModeDoing("Put two lore counters on target Saga you control. The next time one or more enchantment creatures you control enter this turn, each enters with two additional +1/+1 counters on it.",
				TargetPermanent("target Saga you control", OfSubtype("Saga"), YouControl()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					if err := (AddCounter{Target: t.ID, Kind: game.CounterLore, N: 2}).Apply(ctx); err != nil {
						return err
					}
					ctx.Game.GrantNextEntryCountersForEffect(item.Controller, game.CounterPlusOne, 2, true,
						"Storyweave: the next time one or more enchantment creatures you control enter, each enters with two additional +1/+1 counters",
						item.SourceCardID, DurationUntilEndOfTurn(ctx))
					return nil
				}),
		),
	})
}
