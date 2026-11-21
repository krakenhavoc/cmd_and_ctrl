package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaust, Eyes of the Glade — Legendary Creature — Dryad Detective
// {R/W}{G}, 2/2 (Murders at Karlov Manor Commander, #5368):
//
//	"Whenever a creature you control that was turned face up this turn
//	 deals combat damage to a player, draw a card.
//	 {T}: Turn target face-down attacking creature you control face up."
//
// "Turned face up this turn" reads this turn's EventTurnedFaceUp events
// for the damaging creature. The activated ability turns the target up
// for free as an effect (CR 708.8); a cloaked non-creature card refuses
// and stays down (CR 701.40b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af78c13f-67f8-42bf-b012-482a48af4f31",
		Name:         "Kaust, Eyes of the Glade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.Cmdr2S38FaceUpCreatureDealtCombatDamageToAPlayer,
				"Kaust — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Turn target face-down attacking creature you control face up.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target face-down attacking creature you control", effects.FaceDown(), effects.AttackingCreature(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				return effects.TurnFaceUp{Targets: ids}.Apply(ctx)
			},
		}},
	})
}
