package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trickster Mage — Creature — Human Spellshaper {U}, 1/1
// (Nemesis, #5618):
//
//	"{U}, {T}, Discard a card: You may tap or untap target artifact, creature, or land."
//
// The spellshaper cost (nemSpellshaperCost) on Tidal Bore's "you may
// tap or untap" prompt (nemMayTapOrUntapTheTarget): the controller
// picks tap it, untap it or leave it as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5a9fd97-535c-4aa4-9777-6cf8a413b6ec",
		Name:         "Trickster Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: You may tap or untap target artifact, creature, or land.",
			Cost:    effects.NemSpellshaperCost("{U}"),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:  effects.NemMayTapOrUntapTheTarget("Trickster Mage"),
		}},
	})
}
