package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "10821e37-8a83-4d8e-a12f-249a10933a22",
		Name:            "Qasali Pridemage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyChosenPermanent,
		}},
	})
}
