package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SecurityRhoxAltCostKey is the wire key of Security Rhox's {R}{G}
// alternative cost.
const SecurityRhoxAltCostKey = "security-rhox"

// Security Rhox — Creature — Rhino Warrior {2}{R}{G}, 5/4 (#2575):
//
//	"You may pay {R}{G} rather than pay this spell's mana cost. Spend
//	 only mana produced by Treasures to cast it this way."
//
// The {R}{G} is an alternative cost (CR 118.9) with no keyword, declared
// bare as Baleful Mastery's is. Its "spend only mana produced by
// Treasures" restricts that offer alone (AlternativeCost.SpendOnlySources,
// ADR 0040's amendment of 2026-10-22): claimed, the payment takes only
// mana whose recorded source was a Treasure, whether it is floating or
// the auto-tapper cracks the Treasure for it. The printed {2}{R}{G} is
// paid with any mana, as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dffc6785-31b0-487c-b9a0-980bca39753d",
		Name:         "Security Rhox",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:              SecurityRhoxAltCostKey,
			Label:            "Pay {R}{G} with mana from Treasures rather than pay this spell's mana cost",
			ManaCost:         "{R}{G}",
			SpendOnlySources: SpendOnlyManaFromTreasures,
		}},
	})
}
