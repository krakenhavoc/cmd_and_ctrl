package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clockwork Hydra — Artifact Creature — Hydra {5}, 0/0 (Dominaria
// Remastered):
//
//	"This creature enters with four +1/+1 counters on it.
//	 Whenever this creature attacks or blocks, remove a +1/+1 counter from
//	 it. If you do, it deals 1 damage to any target.
//	 {T}: Put a +1/+1 counter on this creature."
//
// The attacks-or-blocks trigger is the Clockwork one (thisAttacksOrBlocks,
// once per combat however many it blocks); its target is chosen as it goes
// on the stack (CR 603.3d). On resolution the counter comes off first and
// the damage is dealt only if it did (CR 608.2c); a Hydra that has left the
// battlefield has no counter to remove, so deals none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fd37ff3-e75d-459a-8717-8f6a9865fb32",
		Name:         "Clockwork Hydra",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 4, "Clockwork Hydra: enters with four +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
					"Clockwork Hydra — remove a +1/+1 counter from it; if you do, it deals 1 damage to any target",
					clockworkHydraPing),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Cost:    effects.TapCost(),
			Effect:  effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}

func clockworkHydraPing(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !effects.OnBattlefield(g, item.SourceCardID) || c.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	if err := effects.RemoveACounterFromThis(game.CounterPlusOne)(g, item); err != nil {
		return err
	}
	return effects.SourceDealsDamageToEachLegalTarget(1)(g, item)
}
