package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost in the Woods — Enchantment {3}{G}{G}:
//
//	"Whenever a creature attacks you or a planeswalker you control, reveal the
//	 top card of your library. If it's a Forest card, remove that creature
//	 from combat. Then put the revealed card on the bottom of your library."
//
// One trigger per attacking creature, as each is declared (CR 508.1m). An
// attack at a battle is not an attack on you (CR 508.3e).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a354b7f3-bfd0-47c6-bb16-1a5f8e26467d",
		Name:         "Lost in the Woods",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.CmdrB1CreatureAttacksYouOrYourPlaneswalker,
				"Lost in the Woods — reveal the top card of your library; if it's a Forest card, remove that creature from combat; then put it on the bottom",
				effects.OldB443LostInTheWoods),
		},
	})
}
