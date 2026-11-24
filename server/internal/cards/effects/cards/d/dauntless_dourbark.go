package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dauntless Dourbark — Creature — Treefolk Warrior {3}{G}, */* (Secret Lair
// Drop):
//
//	"Dauntless Dourbark's power and toughness are each equal to the number of
//	 Forests you control plus the number of Treefolk you control.
//	 This creature has trample as long as you control another Treefolk."
//
// The size is a layer-7a characteristic-defining ability (CR 604.3): each
// Forest you control counts once and each Treefolk you control counts once,
// so Dourbark counts itself as a Treefolk and a Forest Treefolk counts
// twice. The trample is a layer-6 grant that is on only while another
// Treefolk is under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fd25479-b65c-48f7-bb74-95d49af65ac1",
		Name:         "Dauntless Dourbark",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB021ForestsPlusTreefolkPT("Dauntless Dourbark's power and toughness are each equal to the number of Forests you control plus the number of Treefolk you control."),
			effects.S82OldB021TrampleWhileAnotherTreefolk(),
		},
	})
}
