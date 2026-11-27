package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caltrops — {3} Artifact (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"Whenever a creature attacks, this artifact deals 1 damage to it."
//
// One trigger per attacking creature, whoever controls it; the damage is
// dealt only if the creature is still the same object (S82OldB704CaltropsDamage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b13f97b3-56ae-4131-a93b-b162c8ff7cca",
		Name:         "Caltrops",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil
			}, "Caltrops — it deals 1 damage to the attacking creature", effects.S82OldB704CaltropsDamage),
		},
	})
}
