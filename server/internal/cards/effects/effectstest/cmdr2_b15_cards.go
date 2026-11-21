package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func C2b15Forest() game.Card { return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"} }
func C2b15Bear() game.Card {
	return game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
}
func C2b15Spell() game.Card { return game.Card{Name: "Opt", TypeLine: "Instant"} }
