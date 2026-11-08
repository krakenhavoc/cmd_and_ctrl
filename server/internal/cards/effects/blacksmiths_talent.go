package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blacksmith's Talent — Enchantment — Class {R} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create a colorless Equipment artifact token
//	 named Sword with "Equipped creature gets +1/+1" and equip {2}.
//	 {2}{R}: Level 2
//	 At the beginning of combat on your turn, attach target Equipment you
//	 control to up to one target creature you control.
//	 {3}{R}: Level 3
//	 During your turn, equipped creatures you control have double strike
//	 and haste."
//
// ADR 0071 Class machinery (Hunter's Talent): LevelUp for each level, and
// the level-2 trigger and level-3 static gated at their level (CR 716.2a).
//
//   - Level 1 makes Sword, a catalog token (ADR 0083) like Gwyndala's
//     Neuroflux Heirloom: its pump and its equip ability are registered
//     under the token's key, so it survives a restore point. It is created
//     unattached, as printed.
//   - Level 2 is Iron Hills Stalwart's two clauses (AttachClauseTargets) at
//     the beginning of combat on your turn: an Equipment you control must
//     be chosen for it to go on the stack, and with no creature chosen it
//     attaches nothing (CR 701.3b).
//   - Level 3 is Long-Lost Lances' "during your turn, creatures you control
//     that are equipped" as two layer-6 keyword grants.
//
// No simplification.
func init() {
	doubleStrike := Labelled(KeywordGrant(hobB5EquippedCreatureOnYourTurn, "double strike"),
		"During your turn, equipped creatures you control have double strike.")
	doubleStrike.ActiveWhen = Level(3)
	haste := Labelled(KeywordGrant(hobB5EquippedCreatureOnYourTurn, "haste"),
		"During your turn, equipped creatures you control have haste.")
	haste.ActiveWhen = Level(3)
	Register(Spec{
		OracleID:     "9fdad42a-9db5-4d91-9e80-9d5ce19038b8",
		Name:         "Blacksmith's Talent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blacksmith's Talent — create a colorless Equipment artifact token named Sword",
				Do(CreateToken{Template: tokenFromCatalog(printedBlacksmithsSwordToken), N: 1})),
			AtLevel(2, Targeting(
				AtBeginningOfYourCombat("Blacksmith's Talent — attach target Equipment you control to up to one target creature you control",
					AttachClauseTargets),
				Clauses(
					TargetPermanent("target Equipment you control", HasSubtype("Equipment"), YouControl()),
					UpToOneTargetCreature("up to one target creature you control", YouControl()),
				))),
		},
		Static: []game.StaticAbility{doubleStrike, haste},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{2}{R}")),
			LevelUp(3, ManaCost("{3}{R}")),
		},
	})
}

// printedBlacksmithsSwordToken is the catalog's entry for Sword: +1/+1 for
// the equipped creature, and equip {2}.
func printedBlacksmithsSwordToken() tokenTemplate {
	return tokenTemplate{
		Slug: "blacksmiths-talent-sword",
		Card: game.Card{
			Name:     "Sword",
			TypeLine: "Token Artifact — Equipment",
		},
		Text:      "Equipped creature gets +1/+1.\nEquip {2}",
		Static:    []game.StaticAbility{PumpAttached(1, 1)},
		Activated: activatedShapes([]ActivatedAbility{EquipAbility("{2}")}),
	}
}
