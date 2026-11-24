package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burdened Aerialist — Creature — Human Pirate {2}{U}, 3/1 (Modern Horizons 2):
//
//	"When this creature enters, create a Treasure token.
//	 Whenever you sacrifice a token, this creature gains flying until end of
//	 turn."
//
// One trigger per token sacrificed, whether to a cost (a Treasure's own mana
// ability) or an effect; the flying is only for the Aerialist the ability
// came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd39cccb-095c-4ef4-8423-f4d328849295",
		Name:         "Burdened Aerialist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Burdened Aerialist — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			effects.On(game.EventSacrifice, effects.OldB033YouSacrificedAToken,
				"Burdened Aerialist — this creature gains flying until end of turn",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Burdened Aerialist — flying until end of turn")),
		},
	})
}
