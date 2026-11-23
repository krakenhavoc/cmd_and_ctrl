package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hivespine Wolverine — Creature — Elemental Wolverine {3}{G}{G}, 5/4
// (Bloomburrow, #4467):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on target creature you control.
//	 • This creature fights target creature token.
//	 • Destroy target artifact or enchantment."
//
// Coliseum Behemoth's modal entry trigger (a real mode pick, #764) with
// three bullets: Inspired Inventor's counter on a creature you control
// (the Wolverine itself included), Ravager Wurm's fight against any
// creature token, yours or an opponent's, and the Behemoth's destroy.
// The fight deals nothing if either creature has left (CR 701.14b); a
// bullet whose target left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Hivespine Wolverine — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on target creature you control.",
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.PlusOneCounterOnTheModesTarget),
		effects.ModeDoing("This creature fights target creature token.",
			effects.Harmful(effects.TargetCreature("target creature token", effects.IsTokenPredicate())),
			effects.ThisFightsTheModesTarget),
		effects.ModeDoing("Destroy target artifact or enchantment.",
			effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			effects.DestroyTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "312be209-dba5-4e6c-938e-1287642f67d4",
		Name:         "Hivespine Wolverine",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
