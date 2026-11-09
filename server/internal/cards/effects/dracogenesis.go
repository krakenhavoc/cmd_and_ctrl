package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dracogenesis — Enchantment {6}{R}{R} (Tarkir: Dragonstorm, #4463):
//
//	"You may cast Dragon spells without paying their mana costs."
//
// Omniscience's granted free alternative cost (ADR 0118 §3), narrowed to
// spells with the Dragon type (CastDragonSpellsWithoutPayingManaCost) and
// open to every zone a Dragon spell is cast from. It is an alternative cost,
// so a spell cast this way can't also use another one (CR 118.9a); X is 0 and
// mandatory additional costs are still paid (CR 118.9d, the rulings), and a
// Dragon keeps its own timing, so a sorcery-speed one still waits for a main
// phase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "9b06f2bf-146d-4b30-b467-e16f1a942253",
		Name:                    "Dracogenesis",
		Completeness:            CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{CastDragonSpellsWithoutPayingManaCost()},
	})
}
