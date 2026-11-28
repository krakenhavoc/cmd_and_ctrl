package cardsv

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ventifact Bottle — Artifact {3} (Mirage):
//
//	"{X}{1}, {T}: Put X charge counters on this artifact. Activate only as a
//	 sorcery.
//	 At the beginning of your first main phase, if this artifact has a charge
//	 counter on it, tap it and remove all charge counters from it. Add {C} for
//	 each charge counter removed this way."
//
// The counters are put on as the ability resolves. The main-phase trigger has an
// intervening "if" (CR 603.4) checked as the phase begins and again as the
// trigger resolves; it taps the Bottle, removes every counter and adds {C} for
// the number removed, read as the trigger resolves (CR 608.2h). The mana empties
// with the precombat main phase (CR 106.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f167e2f-7b89-4ae2-8332-7c978f71522e",
		Name:         "Ventifact Bottle",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:        "{X}{1}, {T}: Put X charge counters on this artifact. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{X}{1}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: ctx.X()}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "this artifact has a charge counter on it",
				Test: oldB804BottleHasCounter,
			}, effects.AtYourPrecombatMain("Ventifact Bottle — tap it, remove all charge counters, add {C} for each", oldB804BottleRelease)),
		},
	})
}

func oldB804BottleHasCounter(ctx *effects.Context) bool {
	info, ok := ctx.SourcePermanent()
	return ok && !info.Left && info.Counters[game.CounterCharge] > 0
}

// oldB804BottleRelease is the trigger's body: tap the Bottle, remove all its
// charge counters and add {C} for each one removed.
func oldB804BottleRelease(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Left {
		return nil
	}
	n := info.Counters[game.CounterCharge]
	if n <= 0 {
		return nil
	}
	if err := (effects.TapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: -n}).Apply(ctx); err != nil {
		return err
	}
	return effects.AddMana{Produced: strings.Repeat("{C}", n)}.Apply(ctx)
}
