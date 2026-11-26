package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jund Sojourners — Creature — Lizard Shaman {B}{R}{G}, 3/2 (Alara Reborn,
// #5586):
//
//	"When you cycle this card and when this creature dies, you may have it
//	 deal 1 damage to any target.
//	 Cycling {2}{R}"
//
// One ability with two conditions (S82OldB501CycledOrDies). The damage is
// dealt by the card itself, from the graveyard (last-known information,
// CR 608.2h), as Solar Blast's cycle trigger is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84a3ebc2-5c38-40ef-b5f4-62a92cbca493",
		Name:         "Jund Sojourners",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.Optional(
					effects.S82OldB501CycledOrDies("Jund Sojourners — you may have it deal 1 damage to any target",
						func(g *game.Game, item *game.StackItem) error {
							return effects.DamageToFirstTargetStep(1)(item, effects.NewContext(g, item))
						}),
					"Jund Sojourners — have it deal 1 damage to any target?"),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
