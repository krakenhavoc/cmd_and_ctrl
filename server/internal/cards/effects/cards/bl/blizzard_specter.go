package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blizzard Specter — Snow Creature — Specter {2}{U}{B}, 2/3 (Coldsnap):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, choose one —
//	 • That player returns a permanent they control to its owner's hand.
//	 • That player discards a card."
//
// One modal trigger (CR 700.2, CR 603.3c): the mode is chosen as the trigger
// goes on the stack. The damaged player makes the pick in the first mode, from
// the permanents they control (none: nothing happens); in the second they
// discard a card of their choice.
//
// No simplification.
func init() {
	t := effects.WheneverThisDealsCombatDamageToAPlayer("Blizzard Specter — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("That player returns a permanent they control to its owner's hand.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				who := item.Trigger.Event.Target
				if who == uuid.Nil {
					return nil
				}
				return effects.ChoosePermanents{
					Player:   who,
					Question: "Blizzard Specter — return a permanent you control to its owner's hand",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var out []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of {
								out = append(out, c.InstanceID)
							}
						}
						if len(out) == 0 {
							return nil, 0, 0
						}
						return out, 1, 1
					},
					Keyed: effects.BouncedChosen,
				}.Apply(ctx)
			}),
		effects.ModeDoing("That player discards a card.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DiscardOne(ctx.Game, item, item.Trigger.Event.Target)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "b6f4148d-8511-46b3-93ec-2f5eb1e4ffad",
		Name:            "Blizzard Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
