package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squirrel Sanctuary — Enchantment {G} (Modern Horizons 2):
//
//	"When this enchantment enters, create a 1/1 green Squirrel creature token.
//	 Whenever a nontoken creature you control dies, you may pay {1}. If you do,
//	 return this enchantment to its owner's hand."
//
// The second trigger fires once per nontoken creature that dies under your
// control (CR 603.10a). The payment is offered as it resolves; only the
// Sanctuary the ability came from is returned (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "693d20e9-ad92-44ad-9190-d247c94660eb",
		Name:         "Squirrel Sanctuary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Squirrel Sanctuary — create a 1/1 green Squirrel creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Squirrel"), N: 1})),
			effects.On(game.EventLTB, effects.Cmdr2S30NontokenCreatureYouControlDied,
				"Squirrel Sanctuary — you may pay {1}; if you do, return this enchantment to its owner's hand",
				effects.OldB033SquirrelSanctuaryAsk),
		},
	})
}
