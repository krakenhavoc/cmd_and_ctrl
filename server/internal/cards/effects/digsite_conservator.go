package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Digsite Conservator — Artifact Creature — Gnome {2}, 2/1:
//
//	"Sacrifice this creature: Exile up to four target cards from a
//	 single graveyard. Activate only as a sorcery.
//	 When this creature dies, you may pay {4}. If you do, discover 4."
//
// The card #1807 was filed for (ADR 0106 §5). It waited whole on the
// sameness rule: without it the ability could reach four graveyards
// at once, which is stronger than printed. The discover half is ADR
// 0099's, and it was buildable all along.
//
// The dies trigger triggers however the Gnome dies, not only through
// its own ability (the 2023-11-10 ruling). The "may" is the payment:
// MayPay asks on resolution, and discover runs only when {4} was paid.
// Discover exiles cards whatever the player chooses afterwards; only
// casting the card it finds is optional (CR 701.57a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ced2ae56-988a-4c60-93d0-85a2261be72c",
		Name:         "Digsite Conservator",
		Completeness: CompletenessFull,
		Discovers:    true,
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice this creature: Exile up to four target cards from a single graveyard. Activate only as a sorcery.",
			Cost:         SacrificeThis(),
			Targets:      upToNCardsFromASingleGraveyard(4),
			SorcerySpeed: true,
			Effect:       ExileTargetCards,
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Digsite Conservator — you may pay {4} to discover 4", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{4}",
					Question: "Digsite Conservator — pay {4} to discover 4?",
					OnPay: func(ctx *Context) error {
						return Discover{N: 4}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
