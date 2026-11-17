package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Fleshgorger — Artifact Creature — Phyrexian Wurm {7}, 7/5
// (The Brothers' War, #5656): "Prototype {1}{B}{B} — 3/3", "Menace,
// lifelink", "Ward—Pay life equal to this creature's power."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Menace and lifelink are printed keywords. The ward is
// WardLifeEqualToPower (Raubahn's): the life is read as the ward
// resolves, so it is 3 for a prototyped Fleshgorger and 7 otherwise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d3a5a830-cd14-49da-9412-c50049c74c92",
		Name:            "Phyrexian Fleshgorger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLifeEqualToPower(), "Phyrexian Fleshgorger — ward, pay life equal to this creature's power"),
		},
	})
}
