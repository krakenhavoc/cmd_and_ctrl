package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time Bomb — Artifact {4} (Ice Age, S82 OLD batch 845, #5636):
//
//	"At the beginning of your upkeep, put a time counter on this artifact.
//	 {1}, {T}, Sacrifice this artifact: This artifact deals damage equal to
//	 the number of time counters on it to each creature and each player."
//
// Armageddon Clock's upkeep counter and Pyrohemia's damage to every creature
// and every player. The amount is read as the ability resolves from the
// Bomb as it last existed (CR 608.2h): the sacrifice is a cost, so the
// permanent is already gone, and its last-known counters supply the number.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29062094-594d-4557-9496-11a15d7e3bb0",
		Name:         "Time Bomb",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Time Bomb — put a time counter", effects.PutACounterOnThis("time")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: This artifact deals damage equal to the number of time counters on it to each creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := 0
				if info, ok := ctx.SourcePermanent(); ok {
					n = info.Counters["time"]
				}
				if n <= 0 {
					return nil
				}
				return effects.B23DamageEachCreatureAndEachPlayer(ctx, n)
			},
		}},
	})
}
