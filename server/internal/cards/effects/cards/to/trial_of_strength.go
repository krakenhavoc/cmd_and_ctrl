package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trial of Strength — {2}{G} Enchantment (Amonkhet, #5535):
//
//	"When this enchantment enters, create a 4/2 green Beast creature token.
//	 When a Cartouche you control enters, return this enchantment to its
//	 owner's hand."
//
// The second ability is the Trials' shared trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b272cb56-dcba-479e-9fba-5b31e616a95f",
		Name:         "Trial of Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Trial of Strength — create a 4/2 green Beast",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/2 green Beast"), N: 1})),
			effects.WhenACartoucheYouControlEntersReturnThis("Trial of Strength"),
		},
	})
}
