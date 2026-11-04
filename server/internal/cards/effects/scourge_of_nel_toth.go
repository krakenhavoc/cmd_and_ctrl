package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scourge of Nel Toth — Creature — Zombie Dragon {5}{B}{B}, 6/6
// (Jumpstart, #4892):
//
//	"Flying
//	 You may cast this creature from your graveyard by paying {B}{B} and
//	 sacrificing two creatures rather than paying its mana cost."
//
// An alternative cost bound to the graveyard (CR 118.9): {B}{B} plus two
// creatures, sacrificed with the spell already on the stack. From the
// graveyard it can be cast only this way, and from hand only for its mana
// cost. It is still a creature spell, cast at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "32972d7a-c7c6-437c-8e7c-6c7d56b130e4",
		Name:             "Scourge of Nel Toth",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{jmpB2CastFromGraveyardBySacrificing("{B}{B}", 2, "two creatures", Creature())},
	})
}
