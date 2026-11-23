package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gonti, Lord of Luxury — Legendary Creature — Aetherborn Rogue {2}{B}{B},
// 2/3 (Outlaws of Thunder Junction Commander):
//
//	"Deathtouch
//	 When Gonti enters, look at the top four cards of target opponent's
//	 library, exile one of them face down, then put the rest on the bottom
//	 of that library in a random order. You may cast that card for as long
//	 as it remains exiled, and mana of any type can be spent to cast that
//	 spell."
//
// "Cast", not "play": the permission is CastOnly, so an exiled land stays
// exiled. A library of one card exiles it with no choice to make.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d5be0ff6-39d9-4f9f-a028-e62dc38463f2",
		Name:            "Gonti, Lord of Luxury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Gonti, Lord of Luxury — look at the top four cards of target opponent's library, exile one face down, put the rest on the bottom in a random order", effects.Cmdr2S44GontiLook),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
	})
}
