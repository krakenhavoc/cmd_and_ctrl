package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lieutenant Hikaru Sulu — Legendary Creature — Human Pilot {2}{G}, 3/3
// (Star Trek Commander, #4974):
//
//	"Whenever Lieutenant Sulu becomes tapped, he explores. (Reveal the
//	 top card of your library. Put that card into your hand if it's a
//	 land. Otherwise, put a +1/+1 counter on this creature, then put the
//	 card back or put it into your graveyard.)
//	 {2}{U}: Target creature can't be blocked this turn."
//
// "Becomes tapped" is ThisBecameTapped over both tap events (Brad
// Boimler's), so an attack taps him and a vigilance attack does not. The
// explore is the engine's CR 701.44 action on Sulu as the object the
// trigger saw: one that left before it resolves still reveals and gets
// no counter (CR 701.44c). The activated ability is Elvenking's Harper's
// at {2}{U}: any creature, any time you have priority, re-checked as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "565cfa73-8c6b-4296-bcb8-388de8c960ff",
		Name:         "Lieutenant Hikaru Sulu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Lieutenant Hikaru Sulu — he explores",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					ref, _ := ctx.SourceRef()
					return ctx.Game.ExploreForEffect(ctx.Source(), ref, ctx.Controller())
				}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeBlockedThisTurn("Lieutenant Hikaru Sulu"),
		}},
	})
}
