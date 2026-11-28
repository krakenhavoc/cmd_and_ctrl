package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Callous Giant — Creature — Giant {4}{R}{R}, 4/4 (Invasion, S82 #5616):
//
//	"If a source would deal 3 or less damage to this creature, prevent that
//	 damage."
//
// A standing CR 615 prevention for each damage event aimed at this creature,
// combat or not, of 1 to 3: the amount is the one event's, so two 2-damage
// sources are each prevented. Damage of 4 or more is dealt in full.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a2336ff-c8ed-4daa-b182-403c21bc8133",
		Name:         "Callous Giant",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID &&
					ev.DamageAmount > 0 && ev.DamageAmount <= 3
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Callous Giant — prevent damage of 3 or less dealt to it",
		}},
	})
}
