package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mawloc — Creature — Tyranid {X}{R}{G}, 2/2:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is
//	 5 or more, draw a card when it enters.)
//	 Terror from the Deep — When this creature enters, it fights up to
//	 one target creature an opponent controls. If that creature would
//	 die this turn, exile it instead."
//
// Ravenous is CR 702.156a's two halves, written out. "Enters with X
// +1/+1 counters" is the CR 614.1c entry clause (XCounters), so the
// counters are on Mawloc before its enters triggers look at it. "When
// it enters, if X is 5 or more, draw a card" is an intervening-if
// enters trigger (CR 603.4) that reads the X
// announced for the spell (CR 107.3m, Card.CastX). That is the X the
// ruling names, not the number of counters it actually got. A Mawloc
// that entered without being cast has X = 0 and draws nothing.
//
// Terror from the Deep is "up to one" target, so it can be declined.
// The fight is b10Fight: a Mawloc that has left the battlefield fights
// nothing. The exile rider is on "that creature", so a legal target is
// marked whether or not the fight happened (the 2022-10-07 ruling:
// "even if Mawloc has left the battlefield by that time").
//
// No simplifications.
func init() {
	const fight = "Mawloc — Terror from the Deep: it fights up to one target creature an opponent controls"
	effects.Register(effects.Spec{
		OracleID:                   "e5d928dc-b465-4cf7-ab11-d5bd3328f8e7",
		Name:                       "Mawloc",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				// CR 603.4: X is fixed when the spell is cast, so the
				// re-check as the trigger resolves cannot differ.
				return effects.Self(ev, source, lki, g) && source.CastX() >= 5
			}, "Mawloc — ravenous: X was 5 or more, draw a card", func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			}),
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Key:       fight,
				Targets:   effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := effects.B10Fight(ctx, item.SourceCardID, id); err != nil {
						return err
					}
					return effects.ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
				},
			},
		},
	})
}
