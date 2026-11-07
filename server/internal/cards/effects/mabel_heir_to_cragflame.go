package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mabel, Heir to Cragflame — Legendary Creature — Mouse Soldier
// {1}{R}{W}, 3/3 (Bloomburrow, #4467):
//
//	"Other Mice you control get +1/+1.
//	 When Mabel enters, create Cragflame, a legendary colorless
//	 Equipment artifact token with "Equipped creature gets +1/+1 and has
//	 vigilance, trample, and haste" and equip {2}."
//
// The lord is a TribeFilter with Others and YoursOnly. Cragflame is a
// catalog token (ADR 0083) like U.S.Agent's Sturdy Shield: its pump,
// its keyword grant and its equip ability are registered under the
// token's key, so the token survives a restore point. It is created
// unattached, as printed. Being legendary, a second Cragflame from a
// second Mabel puts its controller to the legend rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b11e435-5888-4e3b-a255-db817392163f",
		Name:         "Mabel, Heir to Cragflame",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Mouse"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mabel, Heir to Cragflame — create Cragflame, a legendary Equipment token",
				Do(CreateToken{Template: CragflameEquipmentToken(), N: 1})),
		},
	})
}

// CragflameEquipmentToken is Mabel's legendary Equipment token.
func CragflameEquipmentToken() game.Card { return tokenFromCatalog(printedCragflameToken) }

// printedCragflameToken is the catalog's entry for Cragflame: +1/+1,
// vigilance, trample and haste for the equipped creature, and equip {2}.
func printedCragflameToken() tokenTemplate {
	return tokenTemplate{
		Slug: "cragflame",
		Card: game.Card{
			Name:     "Cragflame",
			TypeLine: "Token Legendary Artifact — Equipment",
		},
		Text: "Equipped creature gets +1/+1 and has vigilance, trample, and haste.\nEquip {2}",
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("vigilance", "trample", "haste"),
		},
		Activated: activatedShapes([]ActivatedAbility{EquipAbility("{2}")}),
	}
}
