package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mul Daya Channelers — Creature — Elf Druid Shaman {1}{G}{G}, 2/2 (Rise of
// the Eldrazi, S82 OLD batch 483, #5581):
//
//	"Play with the top card of your library revealed.
//	 As long as the top card of your library is a creature card, this creature
//	 gets +3/+3.
//	 As long as the top card of your library is a land card, this creature has
//	 "{T}: Add two mana of any one color.""
//
// Vampire Nocturnus's reveal and live read of the top card. The pump is a
// layer 7c static of the source alone; the mana ability is a granted bundle
// (New Horizons' shape) that a layer 6 static gives the creature itself only
// while the top card is a land. Both are re-read whenever the top card
// changes (DependsOnLibraryTop).
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Label:               "As long as the top card of your library is a creature card, this creature gets +3/+3.",
		Layer:               game.Layer7PT,
		SubLayer:            game.SubLayer7C_Modify,
		DependsOnLibraryTop: true,
		AppliesTo:           effects.WhileYouHave(effects.OldB483TopCardIsACreature, effects.SelfCreature),
		Apply:               effects.PumpBy(3, 3),
	}
	const grantKey = "mul-daya-channelers/two-of-one-color"
	const grantText = "{T}: Add two mana of any one color."
	grant := effects.GrantAbilities(effects.WhileYouHave(effects.OldB483TopCardIsALand, effects.SelfCreature), grantKey)
	grant.Label = "As long as the top card of your library is a land card, this creature has \"" + grantText + "\""
	grant.DependsOnLibraryTop = true
	effects.Register(effects.Spec{
		OracleID:          "d3daca53-d40e-4537-a16a-f1e97333bc87",
		Name:              "Mul Daya Channelers",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		Static:            []game.StaticAbility{pump, grant},
		Grants: []effects.AbilityGrant{{
			Key:  grantKey,
			Text: grantText,
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: effects.OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
		}},
	})
}
