package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thornmantle Striker — Creature — Elf Rogue {4}{B}, 4/3:
//
//	"When this creature enters, choose one —
//	 • Remove X counters from target permanent, where X is the number of
//	   Elves you control.
//	 • Target creature an opponent controls gets -X/-X until end of turn,
//	   where X is the number of Elves you control."
//
// Cemetery Desecrator's two bullets with another X: the Elves its controller
// controls as the trigger resolves (CR 608.2h), the Striker itself counted if
// it is still there, a changeling too. The mode and its target are chosen as
// the trigger goes on the stack. Removing counters is exact, not "up to": the
// controller picks which kinds and how many of each, X in all, or every
// counter the permanent has when it has fewer.
//
// No simplification.
func init() {
	elves := func(ctx *effects.Context) int {
		return effects.CountYouControl(ctx, effects.And(effects.Creature(), effects.HasSubtype(ctx.TextSubtype("Elf"))))
	}
	trigger := effects.WhenThisEnters("Thornmantle Striker — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Remove X counters from target permanent, where X is the number of Elves you control.",
			effects.TargetPermanent("target permanent"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				x := elves(ctx)
				if x <= 0 {
					return nil
				}
				return effects.OldB069RemoveCounters.ChooseExactly(ctx, t.ID, nil, x, "Thornmantle Striker — remove counters")
			}),
		effects.ModeDoing("Target creature an opponent controls gets -X/-X until end of turn, where X is the number of Elves you control.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				x := elves(ctx)
				if x <= 0 {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x,
					Label: "Thornmantle Striker — -X/-X until end of turn"}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "39f17e13-018c-4a00-ab25-1ac5a904de63",
		Name:         "Thornmantle Striker",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
