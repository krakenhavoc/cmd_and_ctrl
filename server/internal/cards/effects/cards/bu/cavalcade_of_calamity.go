package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavalcade of Calamity — Enchantment {1}{R} (Ravnica Allegiance):
//
//	"Whenever a creature you control with power 1 or less attacks, this enchantment
//	 deals 1 damage to the player or planeswalker that creature is attacking."
//
// One trigger per qualifying attacker, its power read as it is declared. The
// damage goes to what that creature was declared as attacking; a battle takes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e29be33c-e5b0-428d-a057-b1cd61127557",
		Name:         "Cavalcade of Calamity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.OldB175SmallCreatureYouControlAttacks(1),
				"Cavalcade of Calamity — 1 damage to the player or planeswalker that creature is attacking",
				effects.OldB175HitTheAttackedPlayerOrPlaneswalker(1)),
		},
	})
}
