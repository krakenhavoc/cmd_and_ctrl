package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Hills Blacksmith — Creature — Dwarf Artificer {1}{W}, 1/1 (The Hobbit):
//
//	"Double strike
//	 When this creature enters, create a colorless Equipment artifact token
//	 named Axe with "Equipped creature gets +1/+0" and equip {2}."
//
// The Axe is a token template of its own (token_catalog.go): a colorless
// Artifact — Equipment whose printed abilities, the +1/+0 static and the
// equip {2} ability, are registered under its token key like Bonesplitter's
// are under its oracle id, so they survive a snapshot as the key alone. The
// slug is an on-disk identity and is never renamed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9febdfa8-8af2-4aba-b2e6-cb2baad7e02b",
		Name:            "Iron Hills Blacksmith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Iron Hills Blacksmith — create an Axe Equipment token",
				Do(CreateToken{Template: AxeEquipmentToken(), N: 1})),
		},
	})
}

// AxeEquipmentToken is the Axe as PRINTED: a colorless Equipment artifact
// token named Axe, "Equipped creature gets +1/+0", equip {2}.
func AxeEquipmentToken() game.Card { return tokenFromCatalog(printedAxeEquipmentToken) }

// printedAxeEquipmentToken is the Axe as PRINTED, abilities included. It is
// the catalog's entry for this token (token_catalog.go). Its slug is an
// on-disk identity and is never renamed.
func printedAxeEquipmentToken() tokenTemplate {
	return tokenTemplate{
		Slug: "axe-equipment",
		Card: game.Card{
			Name:     "Axe",
			TypeLine: "Token Artifact — Equipment",
		},
		Text:      "Equipped creature gets +1/+0.\nEquip {2}",
		Static:    []game.StaticAbility{PumpAttached(1, 0)},
		Activated: activatedShapes([]ActivatedAbility{EquipAbility("{2}")}),
	}
}
