package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// noctisAltCostKey is the offer a cast through Noctis's permission claims:
// the artifact's printed mana cost and 3 life.
const noctisAltCostKey = "noctis_prince_of_lucis"

// Noctis, Prince of Lucis — Legendary Creature — Human Noble {1}{W}{U}{B},
// 4/3 (Final Fantasy, #4462):
//
//	"Lifelink
//	 You may cast artifact spells from your graveyard by paying 3 life in
//	 addition to paying their other costs. If you cast a spell this way,
//	 that artifact enters with a finality counter on it."
//
// Lifelink is a printed keyword. The cast is a STANDING graveyard
// permission derived off the battlefield (Leonardo, Sewer Samurai's
// shape: no budget, any number of casts, the card's own timing in
// force) filtered to artifact spells. The three life is the permission's
// LifeInAddition (CastPermission.LifeInAddition, #4462): the offer it
// synthesises charges the artifact's printed mana cost and 3 life, so a
// player with fewer than 3 life cannot claim it (CR 119.4); timing,
// targets and every other cost are the spell's own. The finality counter
// is the permission's entry rider (EntersWithCounterKind, widened to any
// permanent spell by EntersWithCountersPermanent, #4462), put on as the
// artifact enters (CR 614.1c), so an artifact cast this way is exiled
// instead if it would be put into a graveyard from the battlefield (the
// ruling). An artifact land is played, not cast, so it is not opened.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d44bdd2-aa31-415a-be71-d9e98ef12334",
		Name:            "Noctis, Prince of Lucis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter: game.PermissionFilter{
				NonLandOnly: true,
				CardType:    "artifact",
			},
			AltCostKey:                  noctisAltCostKey,
			LifeInAddition:              3,
			EntersWithCounterKind:       game.CounterFinality,
			EntersWithCounters:          1,
			EntersWithCountersPermanent: true,
			Label:                       "Cast an artifact spell from your graveyard, paying 3 life in addition; it enters with a finality counter (Noctis, Prince of Lucis)",
		}},
	})
}
