package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cactusfolk Sureshot — Creature — Plant Mercenary {2}{R}{G}, 4/4
// (Outlaws of Thunder Junction, #4468):
//
//	"Reach
//	 Ward {2}
//	 At the beginning of combat on your turn, other creatures you control
//	 with power 4 or greater gain trample and haste until end of turn."
//
// Ward {2} is the catalog's ward trigger. The combat trigger is Rinoa,
// Angel Wing's: the set is fixed as it resolves (CR 611.2c), so a
// creature that reaches power 4 later in the turn gains nothing. "Other"
// is the source OBJECT (trekB11OtherThanSourceObject): a Sureshot that
// left and came back is a new object and is included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "05f15b7b-28bf-4b63-a4ca-b57712a1a66a",
		Name:            "Cactusfolk Sureshot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Cactusfolk Sureshot — ward {2}"),
			effects.AtBeginningOfYourCombat("Cactusfolk Sureshot — other creatures you control with power 4 or greater gain trample and haste",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl(), effects.PowerGE(4), effects.TrekB11OtherThanSourceObject(item)),
						Keywords: []string{"trample", "haste"},
						Label:    "Cactusfolk Sureshot — trample and haste",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
