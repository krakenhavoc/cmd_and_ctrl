package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sight of the Scalelords — Enchantment {4}{G} (Dragons of Tarkir, #5551):
//
//	"At the beginning of combat on your turn, creatures you control with toughness 4 or greater get +2/+2 and gain vigilance until end of turn."
//
// The creatures chosen are those with toughness 4 or greater as the trigger
// resolves (CR 611.2c), so a creature that falls below 4 afterwards keeps the
// bonus.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f868b9b-42f4-4f74-bce2-3f9c063a572c",
		Name:         "Sight of the Scalelords",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Sight of the Scalelords — creatures you control with toughness 4 or greater get +2/+2 and gain vigilance until end of turn",
				effects.Do(
					effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.ToughnessGE(4)), Power: 2, Toughness: 2, Label: "Sight of the Scalelords — +2/+2"},
					effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.ToughnessGE(4)), Keywords: []string{"vigilance"}, Label: "Sight of the Scalelords — vigilance"})),
		},
	})
}
