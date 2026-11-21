package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legacy Weapon — Legendary Artifact {7}:
//
//	"{W}{U}{B}{R}{G}: Exile target permanent.
//	 If Legacy Weapon would be put into a graveyard from anywhere,
//	 reveal Legacy Weapon and shuffle it into its owner's library
//	 instead."
//
// The activated ability is Karn Liberated's −3 at a five-colour mana
// cost. The second line is Blightsteel Colossus's replacement
// (ShuffleIntoOwnersLibraryInstead, CR 614.1a): from every zone, a
// discard included, shuffled in once it lands; against Rest in Peace
// its controller, or its owner when it has none, chooses (CR 616.1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa1a248d-2f76-4f15-b065-08f299af07b9",
		Name:         "Legacy Weapon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Exile target permanent.",
			Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.ExileFirstTarget,
		}},
		Replacements: []game.ReplacementEffect{
			effects.ShuffleIntoOwnersLibraryInstead("Legacy Weapon: shuffled into its owner's library instead"),
		},
	})
}
