package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broodbirth Viper — Creature — Snake {4}{U}, 3/3 (Commander 2015, #5546):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)
//	 Whenever this creature deals combat damage to a player, you may draw a
//	 card."
//
// Myriad is the shared trigger (myriad.go, CR 702.116). The draw is
// Shadowmage Infiltrator's optional combat-damage draw; each myriad copy is
// a Broodbirth Viper too, so each copy that connects offers a draw of its
// own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8455828-9c61-4ef9-a5f1-1f0cc6fbf446",
		Name:         "Broodbirth Viper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Broodbirth Viper"),
			effects.OptionalDraw(effects.WheneverThisDealsCombatDamageToAPlayer("Broodbirth Viper — you may draw a card",
				effects.Do(effects.DrawCards{N: 1})), "Draw a card?", 1),
		},
	})
}
