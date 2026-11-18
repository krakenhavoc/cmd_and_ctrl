package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6158ab85-a929-472c-9618-8e76d2d4b228",
		Name:         "Promise of Bunrei",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Promise of Bunrei — sacrifice this enchantment; if you do, create four 1/1 Spirits",
				promiseOfBunreiSacrifice),
		},
	})
}

// promiseOfBunreiSacrifice is "sacrifice this enchantment. If you do,
// create four 1/1 colorless Spirit creature tokens". An on-disk key.
var promiseOfBunreiSacrifice = SacrificeThisThen("then/promise-of-bunrei-spirits", func(ctx *Context) error {
	return CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 4}.Apply(ctx)
})
