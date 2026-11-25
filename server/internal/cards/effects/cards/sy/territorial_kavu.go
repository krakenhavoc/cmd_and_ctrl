package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Kavu — Creature — Kavu {R}{G}, */* (Modern Horizons 2, #5501):
//
//	"Domain — Territorial Kavu's power and toughness are each equal to the
//	 number of basic land types among lands you control.
//	 Whenever this creature attacks, choose one —
//	 • Discard a card. If you do, draw a card.
//	 • Exile up to one target card from a graveyard."
//
// Domain is a characteristic-defining ability (CR 604.3, layer 7a) over the
// basic land types among your lands (Territorial Maro's count). The attack
// trigger is modal: the first bullet is the rummage (a discard then a draw
// only if a card was discarded), the second targets up to one card in any
// graveyard.
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Territorial Kavu — choose one", func(*game.Game, *game.StackItem) error { return nil })
	attack.Modes = effects.ChooseOne(
		effects.ModeDoing("Discard a card. If you do, draw a card.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DiscardACardIfYouDoDraw(ctx.Game, item)
			}),
		effects.ModeDoing("Exile up to one target card from a graveyard.",
			effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
			effects.ExileTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "5f0be4ce-898e-4849-ab8f-a85fb05c3017",
		Name:         "Territorial Kavu",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.DomainPT("Domain — Territorial Kavu's power and toughness are each equal to the number of basic land types among lands you control", 1, 1),
		},
		Triggered: []game.TriggeredAbility{attack},
	})
}
