package effectstest

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// isTextWordArgument is isTextWordLiteral in any case: the helpers in
// textWordHelperArgs compare case-insensitively, so "swamp" names the
// word as surely as "Swamp" does.
func IsTextWordArgument(lit string) bool {
	return IsTextWordLiteral(lit) || game.IsBasicLandType(lit) || game.IsCreatureType(lit)
}

// isTextWordLiteral is a colour letter, a basic land type, or (#3984) a
// creature type, each in the spelling a card file writes it: the
// creature types are the engine's vetted CR 205.3m vocabulary
// (game.AllCreatureTypes, generated from the catalog's type lines), in
// its canonical spelling, so a word that merely looks like one ("Arcane",
// "Vehicle", "Food", "Equipment" are not creature types) is never a
// finding.
func IsTextWordLiteral(lit string) bool {
	switch lit {
	case "W", "U", "B", "R", "G":
		return true
	}
	if game.IsBasicLandType(lit) && strings.ToUpper(lit[:1]) == lit[:1] {
		return true
	}
	canon, ok := game.CanonicalCreatureType(lit)
	return ok && canon == lit
}
