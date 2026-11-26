package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wardscale Dragon — {4}{W}{W} Creature — Dragon 4/4 (Fate Reforged):
//
//	"Flying
//	 As long as this creature is attacking, defending player can't cast spells."
//
// A cast restriction read as each spell is announced: while the Dragon is
// attacking, the player it attacks (or the controller of the planeswalker or
// battle it attacks, CR 506.3d) can't cast spells; every other player can.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1fb53b6a-b456-4252-9614-12ac72968113",
		Name:            "Wardscale Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastRestrictions: []game.CastRestriction{{
			Label: "As long as Wardscale Dragon is attacking, defending player can't cast spells.",
			Forbids: func(q game.CastQuery) bool {
				if q.Source.AttackingTarget == uuid.Nil {
					return false
				}
				return effects.DefendingPlayerOfAttacker(q.Game, q.Source) == q.Controller
			},
		}},
	})
}
