package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farrelite Priest — Creature — Human Cleric {1}{W}{W}, 1/3 (Fallen
// Empires, #5745):
//
//	"{1}: Add {W}. If this ability has been activated four or more times
//	 this turn, sacrifice this creature at the beginning of the next end
//	 step."
//
// A filter paid with {1} and no tap, so it can be activated any number of
// times. The sentence after "Add {W}" is the activation-count rider
// (mana_activation_count_rider.go): from the fourth activation of this
// ability on this object this turn, a delayed trigger sacrifices the
// Priest at the beginning of the next end step. The auto-tapper does not
// spend it; the player activates it from the board.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ee18e5f-3d12-489e-b337-29208caaed7a",
		Name:         "Farrelite Priest",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{1}"},
			Produced: "{W}",
			Label:    FarrelitePriestManaLabel,
			Rider: effects.SacrificeAtNextEndStepIfActivatedRider(4, FarrelitePriestManaLabel,
				"Farrelite Priest — sacrifice this creature"),
		}},
	})
}

// farrelitePriestManaLabel is the mana ability's label, and so the key
// its activations are counted under.
const FarrelitePriestManaLabel = "{1}: Add {W}. If this ability has been activated four or more times this turn, sacrifice this creature at the beginning of the next end step"
