package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrieking Specter — Creature — Specter {5}{B}, 2/2 (Starter 1999,
// #5620):
//
//	"Flying
//	 Whenever this creature attacks, defending player discards a card."
//
// Witch-king, Bringer of Ruin's attack trigger with Abyssal
// Nightstalker's payoff: the defending player is the player attacked, or
// the controller of the planeswalker or battle attacked
// (defendingPlayerOfAttack), and they choose the card they discard, so
// the payoff is the discard prompt (CR 701.9a). A defender who has left
// the game discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6adfb5f-a21c-4e86-bd16-a9c1e355af34",
		Name:            "Shrieking Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Shrieking Specter — defending player discards a card", effects.AttackDefenderDiscardsACard),
		},
	})
}
