package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krovikan Horror — Creature — Horror Spirit {3}{B} 2/2 (Alliances, S82 OLD
// batch 819):
//
//	"At the beginning of the end step, if this card is in your graveyard with
//	 a creature card directly above it, you may return this card to your hand.
//	 {1}, Sacrifice a creature: This creature deals 1 damage to any target."
//
// It is the end step of every turn, not only yours. The condition is an
// intervening "if" (CR 603.4); see Death Spark for "directly above".
func init() {
	const label = "Krovikan Horror — you may return this card to your hand"
	end := effects.InterveningIf(effects.OldB819GraveyardCreatureAboveCondition,
		effects.On(game.EventBeginEndStep, effects.AllOf(effects.AnyPlayer, effects.OldB819InGraveyardUnderACreature), label,
			func(g *game.Game, item *game.StackItem) error {
				return effects.OldB819ReturnThisFromGraveyardToHand(effects.NewContext(g, item), item.SourceCardID)
			}))
	end.Zones = []game.ZoneKind{game.ZoneGraveyard}
	end = effects.Optional(end, label+"?")
	effects.Register(effects.Spec{
		OracleID:     "a4c15223-784f-43a7-8626-30f96ccfdcaf",
		Name:         "Krovikan Horror",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{end},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: This creature deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToFirstTargetStep(1)(item, effects.NewContext(g, item))
			},
		}},
	})
}
