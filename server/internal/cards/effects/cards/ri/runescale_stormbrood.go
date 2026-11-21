package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runescale Stormbrood // Chilling Screech — Tarkir: Dragonstorm (#4463),
// an omen card (CR 720, #5175):
//
//	Runescale Stormbrood — Creature — Dragon {3}{R}, 2/4
//	  "Flying
//	   Whenever you cast a noncreature spell or a Dragon spell, this
//	   creature gets +2/+0 until end of turn."
//	Chilling Screech — Instant — Omen {1}{U}
//	  "Counter target spell with mana value 2 or less. (Then shuffle this
//	   card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). The cast trigger is
// Firespitter Whelp's filter; an Omen spell is an instant or sorcery, so
// casting another omen card's Omen half triggers it too. Chilling Screech
// is Highly Illogical's counter without the kicker.
//
// No simplification.

const RunescaleStormbroodOracleID = "af7db15c-93af-4c7e-a028-42181d91a923"

func init() {
	effects.Register(effects.Spec{
		OracleID:        RunescaleStormbroodOracleID,
		Name:            "Runescale Stormbrood",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Noncreature(), effects.HasSubtype("Dragon")),
				"Runescale Stormbrood — +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Runescale Stormbrood — +2/+0")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     RunescaleStormbroodOracleID + "#1",
		Name:         "Chilling Screech",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell with mana value 2 or less", effects.ManaValueLE(2)),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
