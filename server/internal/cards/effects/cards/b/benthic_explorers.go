package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benthic Explorers — Creature — Merfolk Scout {3}{U}, 2/4 (Alliances, #5634):
//
//	"{T}, Untap a tapped land an opponent controls: Add one mana of any type
//	 that land could produce."
//
// A mana ability (CR 605.1a) with an untap cost reaching onto an opponent's
// board (#6208): only a TAPPED land an opponent controls pays (CR 701.26b), it
// is not targeted, and the Explorers' own {T} is a separate component so
// summoning sickness applies. The mana is read off the land the cost untapped
// (PaidCost.UntappedOthers): every type it could produce (CR 106.7), colourless
// included, and the activator picks one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc0518fc-d34b-474a-92c5-2cb013828396",
		Name:         "Benthic Explorers",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:       true,
				TapOthers: effects.UntapTappedAnOpponentControls(1, "a tapped land an opponent controls", effects.Land()),
			},
			ProducedForPaid: effects.OldB815AnyTypeTheUntappedLandCouldProduce,
			Label:           "{T}, Untap a tapped land an opponent controls: Add one mana of any type that land could produce.",
		}},
	})
}
