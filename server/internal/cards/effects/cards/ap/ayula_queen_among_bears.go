package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ayula, Queen Among Bears — Legendary Creature — Bear {1}{G}, 2/2
// (Fallout, #5948):
//
//	"Whenever another Bear you control enters, choose one —
//	 • Put two +1/+1 counters on target Bear.
//	 • Target Bear you control fights target creature you don't control."
//
// A modal trigger (CR 603.3c), Ainok Guide's shape: the mode, and then its
// targets, are chosen as the trigger goes on the stack, so the table can
// respond. The first bullet's Bear is any Bear, an opponent's included. The
// second is Bushwhack's fight, happening only while both targets are still
// legal (CR 701.14b). The trigger is the entry of another Bear under
// Ayula's controller, read as it entered (a changeling counts); Ayula's
// own entry is no other Bear.
//
// No simplification.
func init() {
	enters := effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.CardID == source.InstanceID {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature() && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Bear"))
	}, "Ayula, Queen Among Bears — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put two +1/+1 counters on target Bear.",
			effects.TargetCreature("target Bear", effects.HasSubtype("Bear")),
			effects.TwoPlusOneCountersOnModeTarget),
		effects.ModeDoing("Target Bear you control fights target creature you don't control.",
			effects.Clauses(
				effects.TargetCreature("target Bear you control", effects.YouControl(), effects.HasSubtype("Bear")),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			),
			effects.FightTheModesTargets),
	)
	effects.Register(effects.Spec{
		OracleID:     "a79d2fc8-467f-464d-9e73-2dae4b059628",
		Name:         "Ayula, Queen Among Bears",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
