package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riku of Two Reflections — Legendary Creature — Human Wizard {2}{G}{U}{R},
// 2/2 (Commander 2011, #5574):
//
//	"Whenever you cast an instant or sorcery spell, you may pay {U}{R}. If
//	 you do, copy that spell. You may choose new targets for the copy.
//	 Whenever another nontoken creature you control enters, you may pay {G}{U}.
//	 If you do, create a token that's a copy of that creature."
//
// Two triggers, each with its own "you may pay" asked as it resolves. The spell
// copy is Reflective Golem's (from last-known information once the spell has
// resolved); the creature copy is a token copy of the entering creature
// (CR 707.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f697c78d-7e4f-4320-bfc6-2a25e6d7dc94",
		Name:         "Riku of Two Reflections",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Riku of Two Reflections — you may pay {U}{R} to copy that spell",
				effects.S82OldB455MayPayToCopyThatSpell("{U}{R}", "Riku of Two Reflections — pay {U}{R} to copy that spell?")),
			effects.On(game.EventETB, effects.AnotherNontokenCreatureEnteredUnderYourControl,
				"Riku of Two Reflections — you may pay {G}{U} to create a token that's a copy of that creature",
				effects.S82OldB455MayPayToCopyThatCreature("{G}{U}", "Riku of Two Reflections — pay {G}{U} to create a token copy of that creature?")),
		},
	})
}
