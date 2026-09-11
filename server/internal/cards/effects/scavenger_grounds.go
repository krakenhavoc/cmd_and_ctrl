package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID: "5ece7d03-9ee7-4953-a06e-9d8e41874903",
		Name:     "Scavenger Grounds",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice a Desert: Exile all graveyards.",
			Cost: Plus(ManaCost("{2}"), TapCost(),
				game.AbilityCost{SacrificeOther: sacrificeSpec("a Desert", b02IsDesert)}),
			Effect: b02ExileAllGraveyards,
		}},
	})
}
