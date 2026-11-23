package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade Juggler — Creature — Human Rogue {4}{B}, 3/2 (Ravnica: Clue
// Edition):
//
//	"Spectacle {2}{B} (You may cast this spell for its spectacle cost
//	 rather than its mana cost if an opponent lost life this turn.)
//	 When this creature enters, it deals 1 damage to you and you draw a
//	 card."
//
// Spectacle is the shared alternative cost (CR 702.137a). The trigger
// damages its controller (the Juggler itself is the source) and then draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "34400cf5-b8c4-4cf2-9186-51f12037cc51",
		Name:             "Blade Juggler",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Blade Juggler — 1 damage to you and draw a card",
				effects.Do(effects.DamageController{N: 1}, effects.DrawCards{N: 1})),
		},
	})
}
