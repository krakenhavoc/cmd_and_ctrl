package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graven Abomination — {3} Artifact Creature — Horror 3/1 (Hour of Devastation):
//
//	"Whenever this creature attacks, exile target card from defending
//	 player's graveyard."
//
// The target clause is read off the source's attack (Hive of the Eye Tyrant's
// shape): a card in the graveyard of the player this creature attacks. When it
// attacks a planeswalker or battle, that is the player who controls it (or
// protects it), as the engine's defending player.
func init() {
	attack := effects.WheneverThisAttacks("Graven Abomination — exile target card from defending player's graveyard", effects.OldB238ExileEachLegalTargetCard)
	attack.TargetsFrom = effects.OldB238DefendingGraveyardCardTarget
	effects.Register(effects.Spec{
		OracleID:     "2590801e-0de0-4933-be2d-96a1d3cb4585",
		Name:         "Graven Abomination",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
