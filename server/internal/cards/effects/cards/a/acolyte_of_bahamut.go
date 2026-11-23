package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acolyte of Bahamut — Legendary Enchantment — Background {1}{G} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "The first Dragon spell you cast each
//	 turn costs {2} less to cast.""
//
// A granted cost modifier (ADR 0093, the grant carries CostModifiers): the
// commander creature is the modifier's source, so "you" is whoever controls
// it, and a stolen commander discounts its new controller's spells. "First
// Dragon spell each turn" prices a spell while the controller's per-turn
// Dragon tally (game.CastTally.DragonSpells) is still 0, the tally being
// bumped after a cost is determined (CR 601.2f); a changeling spell is a
// Dragon spell. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const acolyteOfBahamutGrant = "acolyte-of-bahamut/first-dragon-spell-discount"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "38477748-bc57-449c-afe1-c5f34caf8eff",
		Name:         "Acolyte of Bahamut",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: acolyteOfBahamutGrant,
			CostModifiers: []game.CostModifier{
				effects.CostsLess(2, "The first Dragon spell you cast each turn costs {2} less to cast.",
					effects.YourSpell(), effects.FirstDragonSpellThisTurn()),
			},
			Text: "The first Dragon spell you cast each turn costs {2} less to cast.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(acolyteOfBahamutGrant)},
	})
}
