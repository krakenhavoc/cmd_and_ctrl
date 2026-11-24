package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firkraag, Cunning Instigator — Legendary Creature — Dragon {3}{U}{R}, 3/3
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Flying, haste
//	 Whenever one or more Dragons you control attack an opponent, goad
//	 target creature that player controls.
//	 Whenever a creature deals combat damage to one of your opponents, if
//	 that creature had to attack this combat, you put a +1/+1 counter on
//	 Firkraag and you draw a card."
//
// The goad trigger is once per opponent attacked in a declaration
// (OncePerBatchPerPlayer, Seifer's shape), counting only Dragons attacking
// that player themself (CR 508.3e: an attack at a planeswalker or battle is
// not an attack at a player). The target is a creature that player
// controls, goaded by the engine's goad (CR 701.15).
//
// "Had to attack this combat" is the engine's CR 508.1d record (#6098,
// game/had_to_attack.go): the creature carried an attack requirement —
// goaded, or "attacks each combat if able" — as the attack declaration was
// made. A creature goaded only after it attacked never had to. The record
// is a fact about the declaration, so the intervening "if" (CR 603.4)
// cannot change between the trigger and its resolution. Any creature's
// damage counts, whoever controls it, as long as the damaged player is one
// of Firkraag's controller's opponents.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "81dad7c5-5950-4c45-adb0-cf1a15a74250",
		Name:            "Firkraag, Cunning Instigator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatchPerPlayer(effects.Targeting(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B16YouAttackedAPlayer(ev, source, g) && ev.Target != source.Controller &&
					effects.AttackerHasSubtype(g, ev, source.TextSubtype("Dragon"))
			}, "Firkraag, Cunning Instigator — goad target creature that player controls", effects.GoadTarget),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls")))),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CreatureDealtCombatDamageToAnOpponentOf(ev, source.Controller, g) && g.HadToAttackThisCombatForEffect(ev.Source)
			}, "Firkraag, Cunning Instigator — put a +1/+1 counter on Firkraag and draw a card",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}, effects.DrawCards{N: 1})),
		},
	})
}
