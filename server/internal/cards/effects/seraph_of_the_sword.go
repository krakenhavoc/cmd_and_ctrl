package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seraph of the Sword — Creature — Angel {3}{W}, 3/3:
//
//	"Flying
//	 Prevent all combat damage that would be dealt to this creature."
//
// A CR 615 prevention on combat damage whose recipient is the Seraph,
// from any source (Marble Priest's shape without the Wall filter).
func init() {
	Register(Spec{
		OracleID:        "84ee74e9-b3e3-4f47-8b3c-5c1a70f5cad6",
		Name:            "Seraph of the Sword",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
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
			Label: "Seraph of the Sword — prevent all combat damage that would be dealt to it",
		}},
	})
}
