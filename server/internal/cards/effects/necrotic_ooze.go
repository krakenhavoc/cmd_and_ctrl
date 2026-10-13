package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrotic Ooze — Creature — Ooze {2}{B}{B}, 4/3:
//
//	"As long as this creature is on the battlefield, it has all
//	 activated abilities of all creature cards in all graveyards."
//
// A layer-6 borrow over every graveyard (ADR 0155). A card in a graveyard
// has no layered characteristic, so its rows come from its catalog entry;
// the borrow rebuilds on every graveyard change (#1117), so a creature
// that dies gives the Ooze its abilities the same pass and one that is
// exiled takes them away. An ability that works only from another zone
// (cycling, unearth) is not usable from the battlefield (CR 113.6).
//
// No simplification.
func init() {
	const clause = "As long as this creature is on the battlefield, it has all activated abilities of all creature cards in all graveyards."
	Register(Spec{
		OracleID:     "2fdeb920-2e48-4308-9517-743eec219c04",
		Name:         "Necrotic Ooze",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: GraveyardLenders(AllGraveyards, ByController(Creature())),
		})},
	})
}
