package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Headhunter — Creature — Human Knight {B/R}{B/R}{B/R}{B/R}, 2/3
// (Throne of Eldraine, #5516):
//
//	"Menace
//	 {B/R}{B/R}{B/R}, Sacrifice another creature or an artifact: This
//	 creature deals 2 damage to target creature or planeswalker."
//
// Menace is the printed keyword. The cost is Mold Folk's sacrifice (another
// creature or an artifact, never the Headhunter itself) beside three hybrid
// mana, paid with either colour in any mix.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fad0118b-d046-4ee7-87d9-e6799f70a7a8",
		Name:            "Elite Headhunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label: "{B/R}{B/R}{B/R}, Sacrifice another creature or an artifact: This creature deals 2 damage to target creature or planeswalker.",
			Cost: effects.Plus(effects.ManaCost("{B/R}{B/R}{B/R}"),
				effects.SacrificeAnotherN(1, "another creature or an artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
