package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodmad Vampire — Creature — Vampire Berserker {2}{R}, 4/1
// (Shadows over Innistrad, #5543):
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 Madness {1}{R} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Hexgold Slith's combat-damage counter, on the creature only while it is
// still the same object (thisStillHere). Madness is grown by the engine
// from the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8a20fc7-025b-403a-b893-ff5efcbdc9d8",
		Name:         "Bloodmad Vampire",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{R}",
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Bloodmad Vampire — put a +1/+1 counter on it", effects.ThisStillHere(effects.PlusOneCountersOnThis(1))),
		},
	})
}
