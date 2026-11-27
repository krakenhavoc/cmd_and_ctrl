package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stromgald Crusader — Creature — Zombie Knight {B}{B}, 2/1 (Coldsnap, S82 OLD b560):
//
//	"Protection from white
//	 {B}: This creature gains flying until end of turn.
//	 {B}{B}: This creature gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.OldB560Crusader("059ed13a-857c-420a-a127-5f720010b9fd", "Stromgald Crusader", "B", "white"))
}
