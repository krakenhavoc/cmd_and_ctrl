package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Shallow Graves — Enchantment — Aura Curse {2}{B} (Commander 2013,
// S82 older sets):
//
//	"Enchant player
//	 Whenever a player attacks enchanted player with one or more creatures,
//	 that attacking player may create a tapped 2/2 black Zombie creature
//	 token."
//
// One trigger per declaration per attacking player, controlled by that
// player, who is asked and who controls the token. The Zombie enters tapped.
//
// No simplification.
var curseOfShallowGravesMake = effects.MayYes("confirm/curse-of-shallow-graves", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("2/2 black Zombie")).EntersTapped(), N: 1}.Apply(ctx)
})

func init() {
	const label = "Curse of Shallow Graves — the attacking player may create a tapped 2/2 black Zombie"
	effects.Register(effects.Spec{
		OracleID:     "dc66dbed-f979-4ad1-b5f9-08538886167f",
		Name:         "Curse of Shallow Graves",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OldB383WhenAPlayerAttacksEnchantedPlayer(label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Curse of Shallow Graves — create a tapped 2/2 black Zombie creature token?",
					Then:     curseOfShallowGravesMake,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
