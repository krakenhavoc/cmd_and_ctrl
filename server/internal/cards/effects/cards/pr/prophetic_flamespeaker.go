package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophetic Flamespeaker — Creature — Human Shaman {1}{R}{R}, 1/3 (Journey into
// Nyx, S82 OLD, #5558):
//
//	"Double strike, trample
//	 Whenever this creature deals combat damage to a player, exile the top card
//	 of your library. You may play it this turn."
//
// Capricious Sliver's impulse body on the creature's own combat damage. Double
// strike deals damage twice, so it triggers once for each hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b290a59a-5752-46bb-90be-9792b57bd467",
		Name:            "Prophetic Flamespeaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Prophetic Flamespeaker — exile the top card of your library; you may play it this turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.ExileTopCardYouMayPlayThisTurn(ctx.Item, ctx, 0)
				}))),
		},
	})
}
