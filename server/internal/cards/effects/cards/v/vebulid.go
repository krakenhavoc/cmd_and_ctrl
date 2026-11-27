package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vebulid — Creature — Phyrexian Horror 0/0 {B} (Urza's Saga, S82 OLD b724):
//
//	This creature enters with a +1/+1 counter on it.
//	At the beginning of your upkeep, you may put a +1/+1 counter on this creature.
//	When this creature attacks or blocks, destroy it at end of combat.
//
// A 0/0 that enters with one counter (a replacement, CR 614.1c) and may grow each upkeep; one trigger for attacking or blocking schedules its destruction for the end of combat step (Cinder Wall's body), blocking once however many attackers it blocks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a887def5-c9e1-4cbb-8b08-79b7bd808deb",
		Name:         "Vebulid",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "Vebulid: enters with a +1/+1 counter"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Vebulid — put a +1/+1 counter on this creature", effects.PutACounterOnThis(game.CounterPlusOne)),
				"Vebulid — put a +1/+1 counter on it?"),
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				"Vebulid — destroy it at end of combat", effects.S82OldB364DestroyThisAtEndOfCombat("Vebulid — destroy it")),
		},
	})
}
