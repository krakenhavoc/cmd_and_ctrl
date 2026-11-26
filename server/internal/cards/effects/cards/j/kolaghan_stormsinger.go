package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kolaghan Stormsinger — Creature — Human Shaman {R}, 1/1 (Dragons of Tarkir,
// S82 old batch 331):
//
//	"Haste
//	 Megamorph {R} (You may cast this card face down as a 2/2 creature for {3}.
//	 Turn it face up any time for its megamorph cost and put a +1/+1 counter on
//	 it.)
//	 When this creature is turned face up, target creature gains haste until
//	 end of turn."
//
// Megamorph is the shared constructor (morph.go). The face-up trigger is
// Silumgar Assassin's, with Hotfoot Gnome's haste grant: its target is any
// creature (the Stormsinger itself included), chosen as the trigger goes on the
// stack and checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0790c8bd-a094-49b5-8843-e8b02cba145b",
		Name:             "Kolaghan Stormsinger",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{R}")},
		Triggered: []game.TriggeredAbility{
			kolaghanStormsingerTurnedFaceUp(),
		},
	})
}

func kolaghanStormsingerTurnedFaceUp() game.TriggeredAbility {
	t := effects.WhenThisIsTurnedFaceUp("Kolaghan Stormsinger — target creature gains haste until end of turn",
		effects.EbTargetCreatureGainsUntilEOT("haste", "Kolaghan Stormsinger — haste until end of turn"))
	t.Targets = effects.TargetCreature("target creature")
	return t
}
