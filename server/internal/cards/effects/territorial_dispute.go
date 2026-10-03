package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a1785817-f17b-471b-a63b-866e7972df1f",
		Name:         "Territorial Dispute",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Territorial Dispute — sacrifice this unless you sacrifice a land",
				sacrificeThisUnlessYouSacrificeALand("Territorial Dispute")),
		},
		LandPlayRestrictions: []game.LandPlayRestriction{
			PlayersCantPlayLands("Players can't play lands."),
		},
	})
}
