package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// stoneforged_blade_token.go — the Stoneforged Blade, the colorless
// Equipment artifact token Nahiri, the Lithomancer's −10 makes (Commander
// Masters, S82 GAP batch 113, #5948). It lives in the core because the
// token catalog lists it.

// StoneforgedBladeToken is Nahiri's Equipment token.
func StoneforgedBladeToken() game.Card { return TokenFromCatalog(printedStoneforgedBladeToken) }

// printedStoneforgedBladeToken is the catalog's entry for the Stoneforged
// Blade: indestructible itself, +5/+5 and double strike for the equipped
// creature, and equip {0}. Its slug is an on-disk identity.
func printedStoneforgedBladeToken() TokenTemplate {
	return TokenTemplate{
		Slug: "stoneforged-blade",
		Card: game.Card{
			Name:     "Stoneforged Blade",
			TypeLine: "Token Artifact — Equipment",
			Keywords: []string{"indestructible"},
		},
		Text: "Indestructible\nEquipped creature gets +5/+5 and has double strike.\nEquip {0}",
		Static: []game.StaticAbility{
			PumpAttached(5, 5),
			GrantToAttached("double strike"),
		},
		Activated: ActivatedShapes([]ActivatedAbility{EquipAbility("{0}")}),
	}
}
