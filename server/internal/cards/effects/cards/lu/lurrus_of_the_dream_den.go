package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurrus of the Dream-Den — Legendary Creature — Cat Nightmare
// {1}{W/B}{W/B}, 3/2:
//
//	"Companion — Each permanent card in your starting deck has mana value
//	 2 or less. (...)
//	 Lifelink
//	 Once during each of your turns, you may cast a permanent spell with
//	 mana value 2 or less from your graveyard."
//
// The permission is Danitha's shape (one use a turn across every
// permanent type, your turn only) with a mana-value cap judged on the face
// being cast, X as zero (CR 202.3e). A land is played, never cast, so it
// is never offered. It lives only while Lurrus is on the battlefield.
// Companion is the deck check's business (internal/deck companion
// registry, ADR 0172), not a card ability.
//
// No simplification.
func init() {
	limit := 2
	effects.Register(effects.Spec{
		OracleID:        "3bc757c1-3adb-4321-8832-8e1cc9e687f7",
		Name:            "Lurrus of the Dream-Den",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		CastPermissions: []game.CastPermission{{
			Zone:              game.ZoneGraveyard,
			CastOnly:          true,
			Filter:            game.PermissionFilter{NonLandPermanentOnly: true},
			MaxSpellManaValue: &limit,
			PerType:           []string{game.PermissionTypePermanent},
			Timing:            game.TimingYourTurnOnly,
			Label:             "Once during each of your turns, cast a permanent spell with mana value 2 or less from your graveyard (Lurrus of the Dream-Den)",
		}},
	})
}
