package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galea, Kindler of Hope — Legendary Creature — Elf Knight {1}{G}{W}{U}, 4/4
// (Adventures in the Forgotten Realms Commander):
//
//	"Vigilance
//	 You may look at the top card of your library any time.
//	 You may cast Aura and Equipment spells from the top of your library. When
//	 you cast an Equipment spell this way, it gains "When this Equipment
//	 enters, attach it to target creature you control.""
//
// The look is the shared private top-of-library clause (LibraryTopOwner). The
// cast is a standing library permission (CR 401.5) for each of the two
// subtypes, so the Equipment one can carry the rider and the Aura one cannot:
// it adds no price and no timing. The rider is the permission's
// (CastPermission.GainsAbilities, Strago and Relm's seam): the Equipment
// spell's permanent gains the "When this Equipment enters" bundle with no
// duration (CR 611.2a), for as long as it is that object, and the trigger
// targets a creature you control as it goes on the stack, then attaches. An
// Equipment cast from the hand gets nothing, and the permission goes with
// Galea.
//
// No simplification.
const galeaAttachGrant = "galea-kindler-of-hope/attach"

func init() {
	effects.Register(effects.Spec{
		OracleID:          "30c0df75-9822-4016-a52b-d2e69dd58124",
		Name:              "Galea, Kindler of Hope",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"vigilance"},
		LibraryTopVisible: game.LibraryTopOwner,
		Grants: []effects.AbilityGrant{{
			Key: galeaAttachGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.WhenThisEnters("Galea — attach this Equipment to target creature you control", effects.AttachSourceToTarget),
					effects.TargetCreature("target creature you control", effects.YouControl())),
			},
			Text: "When this Equipment enters, attach it to target creature you control.",
		}},
		CastPermissions: []game.CastPermission{
			{
				Zone:             game.ZoneLibrary,
				TopOfLibraryOnly: true,
				CastOnly:         true,
				Filter:           game.PermissionFilter{Subtype: "Aura"},
				Label:            "Cast Aura spells from the top of your library (Galea, Kindler of Hope)",
			},
			{
				Zone:             game.ZoneLibrary,
				TopOfLibraryOnly: true,
				CastOnly:         true,
				Filter:           game.PermissionFilter{Subtype: "Equipment"},
				GainsAbilities:   []string{galeaAttachGrant},
				Label:            "Cast Equipment spells from the top of your library; they gain \"When this Equipment enters, attach it to target creature you control\" (Galea, Kindler of Hope)",
			},
		},
	})
}
