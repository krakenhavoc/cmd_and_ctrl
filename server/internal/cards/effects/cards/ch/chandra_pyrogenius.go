package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Pyrogenius — Legendary Planeswalker — Chandra {4}{R}{R}, starting
// loyalty 5 (Kaladesh, #5539):
//
//	"+2: Chandra deals 2 damage to each opponent.
//	 −3: Chandra deals 4 damage to target creature.
//	 −10: Chandra deals 6 damage to target player or planeswalker and each
//	 creature that player or that planeswalker's controller controls."
//
// The +2 is a footprinted damage to each opponent from Chandra. The −3 is a
// damage-to-the-target ability checked again as it resolves (CR 608.2b). The −10
// is Chandra Nalaar's −8 at six damage (a shared body): the player is the target
// or the controller of the targeted planeswalker, and the damage to the player
// and each of their creatures is one simultaneous event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae2c89c2-66d6-484a-a176-9ea19e00f445",
		Name:            "Chandra, Pyrogenius",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+2: Chandra deals 2 damage to each opponent.",
				Cost:   effects.LoyaltyCost(2),
				Effect: effects.Do(effects.DamageEachOpponent{N: 2}),
			},
			{
				Label:   "−3: Chandra deals 4 damage to target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Effect:  effects.DealDamageToTheTarget(4),
			},
			{
				Label:   "−10: Chandra deals 6 damage to target player or planeswalker and each creature that player or that planeswalker's controller controls.",
				Cost:    effects.LoyaltyCost(-10),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 6)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.OldB264DamageTargetPlayerOrPlaneswalkerAndTheirCreatures(g, item, 6)
				},
			},
		},
	})
}
