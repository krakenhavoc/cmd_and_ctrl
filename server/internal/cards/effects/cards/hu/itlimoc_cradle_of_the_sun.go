package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Itlimoc, Cradle of the Sun — Legendary Land, the back face of
// Growing Rites of Itlimoc (growing_rites_of_itlimoc.go):
//
//	"(Transforms from Growing Rites of Itlimoc.)
//	 {T}: Add {G}.
//	 {T}: Add {G} for each creature you control."
//
// Two mana abilities; the second is Gaea's Cradle's, counted as it is
// activated. It is a land, not a creature, so it taps for mana the
// turn it transforms (CR 302.6 is about creatures).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.GrowingRitesOfItlimocOracleID + "#1",
		Name:         "Itlimoc, Cradle of the Sun",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "{T}: Add {G}",
			},
			{
				Cost:           effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.ProducedPerPermanent("G", effects.MatchCreature),
				Label:          "{T}: Add {G} for each creature you control",
			},
		},
	})
}
