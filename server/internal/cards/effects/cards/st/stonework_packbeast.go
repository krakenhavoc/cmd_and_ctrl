package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stonework Packbeast — Artifact Creature — Beast {2}, 2/1 (Zendikar
// Rising, #5508):
//
//	"Stonework Packbeast is also a Cleric, Rogue, Warrior, and Wizard.
//	 {2}: Add one mana of any color."
//
// The extra creature types are a layer 4 static on the card itself, so it
// fills any one slot of the party (CR 700.8b, which assigns each creature to
// the type that maximises the party), and the mana ability is Prismite's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "782c47c5-b108-415c-9b70-a033ac52956f",
		Name:         "Stonework Packbeast",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB095AlsoPartyTypes("Stonework Packbeast")},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color."),
		},
	})
}
