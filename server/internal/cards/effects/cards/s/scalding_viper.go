package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scalding Viper // Steam Clean — Creature — Elemental Snake {1}{R}, 2/1
// // Sorcery — Adventure {1}{U} (Wilds of Eldraine, #4471):
//
//	Scalding Viper — "Whenever an opponent casts a spell with mana value 3
//	 or less, this creature deals 1 damage to that player."
//	Steam Clean — "Return target nonland permanent to its owner's hand.
//	 (Then exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"); the engine exiles the resolved Adventure with permission to
// cast the creature later. The cast trigger is Solarium Sentry's
// condition (ManaValueLE, the spell's mana value as cast) with
// Mindsparker's damage to the caster, dealt by the Viper as an object so
// a Viper that has left the battlefield still deals it. Steam Clean is
// Petty Theft's bounce without the opponent restriction.
//
// No simplification.
const ScaldingViperOracleID = "12c71296-1243-49e6-a196-8b016a7b635c"

func init() {
	effects.Register(effects.Spec{
		OracleID:     ScaldingViperOracleID,
		Name:         "Scalding Viper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.ManaValueLE(3)),
				"Scalding Viper — 1 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					caster := ctx.Trigger().Event.Actor
					ref, ok := ctx.SourceRef()
					if !ok || caster == uuid.Nil {
						return nil
					}
					return effects.DealDamage{SourceObject: &ref, Target: caster, Amount: 1}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     ScaldingViperOracleID + "#1",
		Name:         "Steam Clean",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:    effects.BounceTheTarget,
	})
}
