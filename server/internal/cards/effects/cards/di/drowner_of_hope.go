package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drowner of Hope — Creature — Eldrazi {5}{U}, 5/5 (Modern Horizons 3
// Commander, #5948):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, create two 1/1 colorless Eldrazi Scion
//	 creature tokens. They have "Sacrifice this token: Add {C}."
//	 Sacrifice an Eldrazi Scion: Tap target creature."
//
// The Scions are the catalog's Eldrazi Scion (b28EldraziScionToken), whose
// sacrifice mana ability is real. The activated ability's cost names any
// Eldrazi Scion you control, a Scion token or not, and it can be paid
// with the mana ability's own sacrifice in response to nothing: the two
// abilities are separate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b3f2e85-5ea4-44d6-86e8-e8ce6d4f98c0",
		Name:            "Drowner of Hope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Drowner of Hope — create two 1/1 colorless Eldrazi Scions",
				effects.Do(effects.CreateToken{Template: effects.B28EldraziScionToken(), N: 2})),
				game.Purpose{Tokens: 2, CreatureTokens: 2}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an Eldrazi Scion: Tap target creature.",
			Cost:    effects.SacrificeN(1, "an Eldrazi Scion", effects.HasSubtype("Eldrazi"), effects.HasSubtype("Scion")),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
