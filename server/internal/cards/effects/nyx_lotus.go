package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyx Lotus — Legendary Artifact {4}:
//
//	"Nyx Lotus enters tapped.
//	 {T}: Choose a color. Add an amount of mana of that color equal to
//	 your devotion to that color."
//
// One colour pick whose amount depends on the colour picked (#742):
// the ProducedFunc computes the controller's devotion to each colour
// at activation and writes it into the produced-mana grammar's
// per-option amounts — "{G4|U2}" is "four green or two blue". A colour
// with zero devotion is left off the list, because choosing it adds
// nothing; with no devotion at all the ability adds nothing and the
// Lotus still taps, as printed. With devotion to exactly one colour the
// "pick" has one option and the grammar expands it into ordinary slots.
//
// Devotion counts the mana symbols in the costs of permanents the
// controller controls, hybrid symbols counting toward each of their
// colours (devotionTo, shared with Gray Merchant).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfb8cdde-11f6-441a-942b-32d8c191fc90",
		Name:         "Nyx Lotus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			ProducedFunc:            nyxLotusProduced,
			Label:                   "Choose a color: add mana of that color equal to your devotion to it",
			IgnoreCommanderIdentity: true,
		}},
	})
}

func nyxLotusProduced(g *game.Game, controller, _ uuid.UUID) string {
	var parts []string
	for _, c := range game.AllColors {
		if n := devotionTo(g, controller, c); n > 0 {
			parts = append(parts, c+strconv.Itoa(n))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "{" + strings.Join(parts, "|") + "}"
}
