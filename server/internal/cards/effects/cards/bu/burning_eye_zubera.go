package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning-Eye Zubera — Creature — Zubera Spirit {2}{R}{R}, 3/3 (Saviors of
// Kamigawa, S82 OLD b583):
//
//	"When this creature dies, if 4 or more damage was dealt to it this turn,
//	 this creature deals 3 damage to any target."
//
// Tangled Colony's read of the damage dealt to it this turn: the damage marked
// on the Zubera as it last existed (game.PermanentInfo.DamageMarked, the
// departure record, CR 608.2h). The "if" is intervening (CR 603.4): it is read
// as the Zubera dies, so no trigger and no target without it, and again as the
// ability resolves. The damage is the Zubera's, in the graveyard by then, so
// its last-known information is the source (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24bf8773-c8bb-494c-a0af-aef80f5b3d47",
		Name:         "Burning-Eye Zubera",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.On(game.EventLTB, effects.AllOf(effects.ThisDied, burningEyeZuberaTookFour),
				"Burning-Eye Zubera — 3 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					if !burningEyeZuberaDamaged(g, item.SourceCardID) {
						return nil
					}
					return effects.SourceDealsDamageToEachLegalTarget(3)(g, item)
				}),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}

func burningEyeZuberaDamaged(g *game.Game, id uuid.UUID) bool {
	info, ok := g.LastKnownPermanentForEffect(id)
	return ok && info.DamageMarked >= 4
}

func burningEyeZuberaTookFour(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return burningEyeZuberaDamaged(g, source.InstanceID)
}
