package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stratus Dancer — Creature — Djinn Monk {1}{U}, 2/1 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 Megamorph {1}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
//	 When this creature is turned face up, counter target instant or sorcery spell."
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The face-up trigger counters the first target still legal as it
// resolves (counterTheChosenSpell); with no instant or sorcery spell to
// target it is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "5b6a821c-51e4-44dc-8353-f4d10eab7fcb",
		Name:             "Stratus Dancer",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Megamorph("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisIsTurnedFaceUp("Stratus Dancer — counter target instant or sorcery spell", counterTheChosenSpell),
				instantOrSorcerySpell("target instant or sorcery spell")),
		},
	})
}
