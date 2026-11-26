package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River Serpent — {5}{U} Creature — Serpent 5/5 (Amonkhet, S82 old batch 245):
//
//	"This creature can't attack unless there are five or more cards in your
//	 graveyard.
//	 Cycling {U}"
//
// The restriction is game.AttackTargetRestriction.ControllerGraveyardAtLeast,
// read live at declaration (CR 508.1c) against the controller's own graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6f8c69f-e603-41f5-8f25-da8a102f7c8d",
		Name:         "River Serpent",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CantAttackUnlessYouHaveCardsInGraveyard(5)},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{U}")},
	})
}
