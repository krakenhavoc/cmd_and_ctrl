package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resounding Wave — Instant {2}{U} (Shards of Alara):
//
//	"Return target permanent to its owner's hand.
//	 Cycling {5}{W}{U}{B} ({5}{W}{U}{B}, Discard this card: Draw a card.)
//	 When you cycle this card, return two target permanents to their owners'
//	 hands."
//
// The spell is a targeted bounce of any permanent. The cycle trigger is
// Choking Tethers' graveyard trigger (CR 702.29c) with a two-target clause, two
// different permanents (CR 115.3), each returned if it is still a legal target
// as the trigger resolves (CR 608.2b); a trigger that cannot find two permanents
// to target is removed from the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e607fc7-ec22-4249-8936-70a026f9996d",
		Name:         "Resounding Wave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{5}{W}{U}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.OldB508CycleTrigger("Resounding Wave — cycled: return two target permanents to their owners' hands",
				effects.TargetPermanent("two target permanents").WithCount(2, 2),
				effects.BounceEachLegalTarget),
		},
	})
}
