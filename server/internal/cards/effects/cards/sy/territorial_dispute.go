package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Dispute — Enchantment {4}{R}{R}:
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless
//	 you sacrifice a land.
//	 Players can't play lands."
//
// ADR 0109 §4 (#1895). "Players can't play lands" is a
// LandPlayRestriction that binds every player, its controller included
// (CR 101.2: "can't" beats any land drop, an extra one included). The
// upkeep tax is The Gitrog Monster's, one body shared with it
// (sacrifice_unless_land.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1785817-f17b-471b-a63b-866e7972df1f",
		Name:         "Territorial Dispute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Territorial Dispute — sacrifice this unless you sacrifice a land",
				effects.SacrificeThisUnlessYouSacrificeALand("Territorial Dispute")),
		},
		LandPlayRestrictions: []game.LandPlayRestriction{
			effects.PlayersCantPlayLands("Players can't play lands."),
		},
	})
}
