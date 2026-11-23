package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Everdawn Champion — Creature — Human Soldier {1}{W}{W}, 2/2 (Rivals of
// Ixalan, #5531):
//
//	"Prevent all combat damage that would be dealt to this creature."
//
// Seraph of the Sword's standing prevention (CR 615.1a) without the
// flying: combat damage dealt to it is prevented, and damage from any
// other source is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "964e065f-380d-491b-9225-244c5f9edda7",
		Name:         "Everdawn Champion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.IsCombatDamage && ev.DamageTarget == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Everdawn Champion — prevent all combat damage that would be dealt to it",
		}},
	})
}
