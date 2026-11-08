package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormscale Scion — Creature — Dragon {4}{R}{R}, 4/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Other Dragons you control get +1/+1.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. Copies become tokens.)"
//
// Flying is a printed keyword. The anthem is the tribal lord static
// (TribeFilter, Others and YoursOnly), with "Dragon" read through the
// card's text (CR 612.2). Storm is the keyword constructor every storm
// card uses (CR 702.40a): the count is every spell cast before it this
// turn, any player's, and each copy is a spell copy (CR 707.10) that
// becomes a token as it resolves (CR 111.13), the way Hugh, First of
// Many's replicated artifact spells do. Each copy is a Stormscale Scion
// too, so every Scion on the battlefield pumps the others.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf37426d-b0ad-42c3-81be-79701aa4e0cf",
		Name:            "Stormscale Scion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Dragon"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{Storm()},
	})
}
