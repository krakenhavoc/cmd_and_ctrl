package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenger Grounds — Land — Desert (EDHREC rank 300):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice a Desert: Exile all graveyards."
//
// Bojuka Bog at instant speed for every graveyard at once, on a
// land that taps for mana until the moment it is needed. The
// sacrifice cost is "a Desert" — usually itself, which is why the
// cost is SacrificeOther with a Desert predicate rather than
// SacrificeThis: the Grounds is a Desert, so it is always a legal
// choice, and a deck with a second Desert may keep the Grounds and
// feed it the other one, exactly as printed. The predicate reads the
// post-layer subtype.
//
// Exile-all-graveyards is Farewell's fourth mode: every seat's
// pile, the controller's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ece7d03-9ee7-4953-a06e-9d8e41874903",
		Name:         "Scavenger Grounds",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a Desert: Exile all graveyards.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost: effects.Plus(effects.ManaCost("{2}"), effects.TapCost(),
				game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Desert", effects.B02IsDesert)}),
			Effect: effects.B02ExileAllGraveyards,
		}},
	})
}
