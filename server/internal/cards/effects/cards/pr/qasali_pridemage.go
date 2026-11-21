package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qasali Pridemage — Creature — Cat Wizard {G}{W}, 2/2 (Alara Reborn,
// #5586):
//
//	"Exalted (Whenever a creature you control attacks alone, that creature
//	 gets +1/+1 until end of turn.)
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Exalted is the canonical keyword (CR 702.83a). The sacrifice row is
// Cathar Commando's: the sacrifice is a cost, so the Pridemage is gone
// before the ability resolves, and a target that left in response is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10821e37-8a83-4d8e-a12f-249a10933a22",
		Name:            "Qasali Pridemage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
