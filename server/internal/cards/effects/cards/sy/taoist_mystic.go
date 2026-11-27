package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Taoist Mystic — Creature — Human Mystic {2}{G}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"This creature can't be blocked by creatures with horsemanship."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e705ec38-b8f7-4f99-8cb9-8448dc3e50ec",
		Name:         "Taoist Mystic",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.HasKeyword("horsemanship"), "creatures with horsemanship"),
		},
	})
}
