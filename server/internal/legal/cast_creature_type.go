package legal

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cast_creature_type.go — the enumerator's half of a creature type named
// as a spell is cast (#6245, Selective Snare).

// castCreatureTypeWire is the move's chosen_creature_type: the type
// stamped on a spell that names one, empty on every other.
func castCreatureTypeWire(card game.Card) string {
	if !game.ChoosesCreatureTypeAsCast(card) {
		return ""
	}
	return card.NamedTribe
}
