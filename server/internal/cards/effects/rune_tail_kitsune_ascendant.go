package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rune-Tail, Kitsune Ascendant // Rune-Tail's Essence — {2}{W}
// Legendary Creature — Fox Monk 2/2 // Legendary Enchantment, a flip
// card (#1892, ADR 0151):
//
//	Top:     "When you have 30 or more life, flip Rune-Tail."
//	Flipped: "Prevent all damage that would be dealt to creatures you
//	          control."
//
// The flip is a CR 603.8 state trigger over its controller's life total
// (ADR 0107 §1), so it triggers however the life arrives and, if it is
// countered or the life drops back below 30 before it resolves, once
// more the next time the state holds. It resolves even if the life
// has dropped since: the condition is not rechecked (CR 603.8 is not an
// intervening "if"). FlipThis does nothing to a permanent that has
// already flipped (CR 710.4) or a new object (#1432).
//
// The Essence keeps the card's white colour and {2}{W} cost (CR
// 710.1c) and prevents damage to every creature its controller
// controls, from any source, combat or not.
//
// No simplification.
const runeTailOracleID = "710f1f8a-e5f7-4889-9f12-2611f42d5c73"

func init() {
	Register(Spec{
		OracleID:     runeTailOracleID,
		Name:         "Rune-Tail, Kitsune Ascendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenState("Rune-Tail, Kitsune Ascendant — flip it",
				func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
					p := g.PlayerByIDForEffect(controller)
					return p != nil && !p.Eliminated && p.Life >= 30
				},
				Do(FlipThis{})),
		},
	})
	Register(Spec{
		OracleID:     runeTailOracleID + "#1",
		Name:         "Rune-Tail's Essence",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventAllDamageDealtTo(PreventionStatic{
				To:    ToCreaturesYouControl,
				Label: "Rune-Tail's Essence — prevent all damage that would be dealt to creatures you control",
			}),
		},
	})
}
