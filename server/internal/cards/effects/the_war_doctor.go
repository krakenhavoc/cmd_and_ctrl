package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The War Doctor — Legendary Creature — Time Lord Doctor {2}{R}{W}, 3/5:
//
//	"Whenever one or more other permanents phase out and whenever one or
//	 more other cards are put into exile from anywhere, put a time
//	 counter on The War Doctor.
//	 Whenever The War Doctor attacks, it deals damage equal to the number
//	 of time counters on it to any target. If a creature dealt damage
//	 this way would die this turn, exile it instead."
//
// The first ability has two trigger events, so it is written as two
// once-per-batch triggers under one label each. If both happen in the
// same batch, they are two different trigger events and each one
// triggers (CR 603.2c). Each puts a counter on The War Doctor only
// while it is still the same object on the battlefield.
//
//   - Phasing out is EventPhaseOut. The engine moves a whole phase-out
//     set before it announces any of it, so a War Doctor that phases out
//     with the others is already gone when the events are read, and
//     does not trigger (the 2023-10-13 ruling). Under CR 603.10b it
//     could look back and trigger, but it would then put no counter on
//     a phased-out permanent, so the result is the same.
//   - "Put into exile from anywhere" (otherCardPutIntoExile) is any move
//     of a card into exile, from the battlefield, the stack, a graveyard,
//     a hand or a library. A token is not a card.
//
// The attack trigger reads the counters as it resolves, from the last
// time The War Doctor was on the battlefield if it has left (the
// 2023-10-13 ruling). "A creature dealt damage this way" is registered
// from the damage's continuation, only on a creature that was dealt more
// than 0 damage (ADR 0108 §1 decision 3).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d97553bf-6763-4a7b-8d82-1b438a22aa62",
		Name:         "The War Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventPhaseOut, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != source.InstanceID
			}, "The War Doctor — other permanents phased out: a time counter", theWarDoctorTimeCounter)),
			OncePerBatch(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventCounterSpell},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return otherCardPutIntoExile(ev, source, g)
				},
				Key:    "The War Doctor — other cards put into exile: a time counter",
				Effect: theWarDoctorTimeCounter,
			}),
			{
				Watches:   []game.EventKind{game.EventAttack},
				AppliesTo: ThisAttacked,
				Key:       "The War Doctor — damage equal to its time counters to any target",
				Targets:   TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					n := info.Counters[game.CounterTime]
					for _, t := range ctx.LegalTargets() {
						return DealDamageThen(ctx, t.ID, n, ExileIfDealtDamageWouldDie(item, false))
					}
					return nil
				},
			},
		},
	})
}

// theWarDoctorTimeCounter puts a time counter on The War Doctor, if it
// is still on the battlefield. AddCounter refuses a new object (#1432).
func theWarDoctorTimeCounter(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: 1}.Apply(NewContext(g, item))
}
