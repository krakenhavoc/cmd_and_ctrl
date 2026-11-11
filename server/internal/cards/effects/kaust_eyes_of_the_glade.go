package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "af78c13f-67f8-42bf-b012-482a48af4f31",
		Name:         "Kaust, Eyes of the Glade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, cmdr2S38FaceUpCreatureDealtCombatDamageToAPlayer,
				"Kaust — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Turn target face-down attacking creature you control face up.",
			Cost:    TapCost(),
			Targets: TargetCreature("target face-down attacking creature you control", FaceDown(), AttackingCreature(), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				return TurnFaceUp{Targets: ids}.Apply(ctx)
			},
		}},
	})
}
