package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ced2ae56-988a-4c60-93d0-85a2261be72c",
		Name:         "Digsite Conservator",
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice this creature: Exile up to four target cards from a single graveyard. Activate only as a sorcery.",
			Cost:         effects.SacrificeThis(),
			Targets:      effects.UpToNCardsFromASingleGraveyard(4),
			SorcerySpeed: true,
			Effect:       effects.ExileTargetCards,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Digsite Conservator — you may pay {4} to discover 4", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{4}",
					Question: "Digsite Conservator — pay {4} to discover 4?",
					Then:     digsiteConservatorDiscover,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// digsiteConservatorDiscover is the paid half, keyed (#4264): the
// controller rides the asking item, so a table waiting on the payer is a
// restore point.
var digsiteConservatorDiscover = effects.PayUnlessBranches("pay-unless/digsite-conservator-discover",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.Discover{N: 4}.Apply(ctx) }, nil)
