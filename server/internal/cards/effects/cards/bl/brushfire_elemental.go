package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brushfire Elemental — Creature — Elemental {R}{G}, 1/1 (Zendikar
// Rising, #5508):
//
//	"Haste
//	 This creature can't be blocked by creatures with power 2 or less.
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Haste is printed. The evasion is Steel Leaf Champion's block rule, read
// against each would-be blocker's power as blocks are declared. The
// landfall pump is Saurian Explorer's (an ability word, CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b33a7ec6-b22b-41cf-9312-65add1529b51",
		Name:            "Brushfire Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Brushfire Elemental — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Brushfire Elemental — +2/+2 until end of turn")),
		},
	})
}
