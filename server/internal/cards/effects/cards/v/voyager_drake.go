package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voyager Drake — Creature — Drake {3}{U} 3/3 (Worldwake, S82 OLD batch 489,
// #5582):
//
//	"Multikicker {U}
//	 Flying
//	 When this creature enters, up to X target creatures gain flying until
//	 end of turn, where X is the number of times this creature was kicked."
//
// The target count is read from the permanent's own kicked record as the
// trigger is put on the stack (Batroc the Leaper's shape); with no kicker no
// trigger target is asked.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Voyager Drake — up to X target creatures gain flying until end of turn", effects.S82OldB489TargetsGainFlying)
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		x := game.CardKickedTimes(*source)
		if x <= 0 {
			return nil
		}
		return effects.TargetCreature("up to X target creatures").WithCount(0, x)
	}
	effects.Register(effects.Spec{
		OracleID:        "b2e68a8c-cf07-4c2c-8e23-7602a8ccf831",
		Name:            "Voyager Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Multikicker("{U}", 20)},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
