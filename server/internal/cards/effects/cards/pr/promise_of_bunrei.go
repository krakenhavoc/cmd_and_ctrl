package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Promise of Bunrei — Enchantment {2}{W}:
//
//	"When a creature you control dies, sacrifice this enchantment. If you
//	 do, create four 1/1 colorless Spirit creature tokens."
//
// Bastion of Remembrance's dies-trigger; the sacrifice is mandatory and
// the Spirits follow only if the enchantment really went (The Misty
// Mountains Cold's SacrificeThisThen).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6158ab85-a929-472c-9618-8e76d2d4b228",
		Name:         "Promise of Bunrei",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		Triggered: []game.TriggeredAbility{
			effects.WheneverACreatureYouControlDies("Promise of Bunrei — sacrifice this enchantment; if you do, create four 1/1 Spirits",
				promiseOfBunreiSacrifice),
		},
	})
}

// promiseOfBunreiSacrifice is "sacrifice this enchantment. If you do,
// create four 1/1 colorless Spirit creature tokens". An on-disk key.
var promiseOfBunreiSacrifice = effects.SacrificeThisThen("then/promise-of-bunrei-spirits", func(ctx *effects.Context) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: 4}.Apply(ctx)
})
