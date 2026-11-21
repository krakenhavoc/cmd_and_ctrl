package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "94389047-4976-4be2-aaca-25dcdc538abe",
		Name:         "Storyweave",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Put two +1/+1 counters on target creature you control.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.TwoPlusOneCountersOnModeTarget),
			effects.ModeDoing("Put two lore counters on target Saga you control. The next time one or more enchantment creatures you control enter this turn, each enters with two additional +1/+1 counters on it.",
				effects.TargetPermanent("target Saga you control", effects.OfSubtype("Saga"), effects.YouControl()),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterLore, N: 2}).Apply(ctx); err != nil {
						return err
					}
					ctx.Game.GrantNextEntryCountersForEffect(item.Controller, game.CounterPlusOne, 2, true,
						"Storyweave: the next time one or more enchantment creatures you control enter, each enters with two additional +1/+1 counters",
						item.SourceCardID, effects.DurationUntilEndOfTurn(ctx))
					return nil
				}),
		),
	})
}
