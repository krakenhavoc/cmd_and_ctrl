package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volrath's Stronghold — Legendary Land (Stronghold, S82 OLD b751, #5627):
//
//	"{T}: Add {C}.
//	 {1}{B}, {T}: Put target creature card from your graveyard on top of your
//	 library."
//
// Haunted Crossroads' put (PutChosenTargetOnTopOfLibrary) behind a tap
// activation, beside a colourless mana ability. A target that has left the
// graveyard by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73b8cf90-3c71-4f8b-a29f-61894b7f27c9",
		Name:         "Volrath's Stronghold",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Put target creature card from your graveyard on top of your library.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect:  effects.PutChosenTargetOnTopOfLibrary,
		}},
	})
}
