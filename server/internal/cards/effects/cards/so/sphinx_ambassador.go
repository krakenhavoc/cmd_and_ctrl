package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx Ambassador — Creature — Sphinx {5}{U}{U}, 5/5 (Magic 2010, #5585):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, search that
//	 player's library for a card, then that player chooses a card name. If
//	 you searched for a creature card that doesn't have that name, you may
//	 put it onto the battlefield under your control. Then that player
//	 shuffles."
//
// Bribery's search of another player's library, with the found card held
// in place (HoldFound) while the searched player names a card (CR 201.4);
// a creature card whose name differs may then be put onto the battlefield
// under your control, and the searched player's library is shuffled
// whichever way it went.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69e1a166-02f7-43f8-b099-dbb01034370c",
		Name:            "Sphinx Ambassador",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Sphinx Ambassador — search that player's library for a card", effects.OldB499SphinxSearch),
		},
	})
}
