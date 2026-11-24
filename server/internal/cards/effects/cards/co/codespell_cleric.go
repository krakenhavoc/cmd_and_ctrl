package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Codespell Cleric — Creature — Human Cleric {W}, 1/1 (Kaldheim, #5505):
//
//	"Vigilance
//	 When this creature enters, if it was the second spell you cast this
//	 turn, put a +1/+1 counter on target creature."
//
// An intervening "if" (CR 603.4) about a fixed fact of the cast: the
// ordinal its spell had when it was cast (CastProvenance.CastOrdinal), so a
// third spell cast in response does not change it and a Cleric that was
// not cast (a reanimated or flickered one) never triggers. It cannot stop
// being true while the trigger waits, and a Cleric that has left by
// resolution still was the second spell, so the check at trigger time is
// the whole of it.
func init() {
	t := effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisWasTheNthSpellYouCast(2)),
		"Codespell Cleric — put a +1/+1 counter on target creature", effects.B36CounterOnChosenAnimal)
	t.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:        "4388af80-118f-4bf4-8fd7-e68c58398ee8",
		Name:            "Codespell Cleric",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
