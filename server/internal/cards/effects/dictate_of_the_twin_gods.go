package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dictate of the Twin Gods — Enchantment {3}{R}{R} (EDHREC rank
// 2354):
//
//	"Flash
//	 If a source would deal damage to a permanent or player, it
//	 deals double that damage to that permanent or player instead."
//
// Furnace of Rath at instant speed. Angrath's Marauders' doubler
// with the "you control" clause dropped: every source, everyone's
// damage, the controller's own included — combat damage, a Bolt, a
// pinger, a fight. Flash is the printed keyword the cast path reads
// off PrintedKeywords for a card in hand. CR 616: with a second
// doubler the affected player orders them, and doubling twice is x4
// either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d960d33-623a-4415-ae00-f8cffbc15f5a",
		Name:            "Dictate of the Twin Gods",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageAmount > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Dictate of the Twin Gods: double damage",
		}},
	})
}
