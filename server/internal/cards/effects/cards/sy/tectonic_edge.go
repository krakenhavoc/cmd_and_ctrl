package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tectonic Edge — Land (EDHREC rank 3710):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Destroy target nonbasic land.
//	 Activate only if an opponent controls four or more lands."
//
// Strip Mine that cannot answer the turn-two Cabal Coffers. The gate is
// the ability's activation condition (CR 602.1b, #743):
// OpponentControlsAtLeast reads each opponent on their own, so three
// opponents on two lands apiece do not open it. It is checked once, as
// the ability is activated; an opponent who sacrifices a land in
// response does not stop the destruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4927150d-7ff6-4232-b20e-d2ea245ac710",
		Name:         "Tectonic Edge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}, {T}, Sacrifice this land: Destroy target nonbasic land. Activate only if an opponent controls four or more lands.",
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets:   effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B03Nonbasic()),
			Condition: effects.OpponentControlsAtLeast(4, effects.MatchLand),
			Effect:    effects.DestroyChosenPermanent,
		}},
	})
}
