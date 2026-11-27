package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Searchlight — {3} Artifact.
//
//	{T}: Choose a player. That player adds one mana of any color they
//	choose.
//
// Choosing a player is not targeting (CR 115.10a), so this is a mana
// ability (CR 605.1a) that resolves as it is activated (CR 605.3b). The
// activator names the player with the activation; anyone other than
// the activator is asked for the colour, and the mana enters their
// pool (#6101, game.ManaAbilityShape.ChosenPlayerAdds). The auto-tapper
// never plans it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "511b1174-8ef7-4d17-93de-01b9b81bda6a",
		Name:         "Spectral Searchlight",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:             effects.ManaAbilityCost{Tap: true},
			Produced:         "{W|U|B|R|G}",
			Label:            "Choose a player. That player adds one mana of any color they choose",
			ChosenPlayerAdds: true,
		}},
	})
}
