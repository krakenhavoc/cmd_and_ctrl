package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virulent Emissary — Creature — Elf Assassin {G}, 1/1 (EDHREC rank
// 3343):
//
//	"Deathtouch
//	 Whenever another creature you control enters, you gain 1 life."
//
// A one-drop Soul Warden with deathtouch. Corpse Knight's condition
// (b13AnotherCreatureYouControlEntered — post-layer type, so an
// animated land counts, and a token's entry counts) and one life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d4bec90-7bbc-4385-a2b8-303c7a8d0a0f",
		Name:            "Virulent Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl("Virulent Emissary — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
