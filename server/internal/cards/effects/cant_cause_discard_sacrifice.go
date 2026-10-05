package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// CantBeMadeToSacrifice is "Spells and abilities your opponents control
// can't cause you to sacrifice permanents." (Sigarda, Host of Herons,
// Tajuru Preserver).
func CantBeMadeToSacrifice() []game.OpponentEffectProtection {
	return []game.OpponentEffectProtection{{
		Label:     "Spells and abilities your opponents control can't cause you to sacrifice permanents.",
		Sacrifice: true,
	}}
}

// CantBeMadeToDiscardOrSacrifice is "Spells and abilities your opponents
// control can't cause you to discard cards or sacrifice permanents."
// (Tamiyo, Collector of Tales).
func CantBeMadeToDiscardOrSacrifice() []game.OpponentEffectProtection {
	return []game.OpponentEffectProtection{{
		Label:     "Spells and abilities your opponents control can't cause you to discard cards or sacrifice permanents.",
		Discard:   true,
		Sacrifice: true,
	}}
}
