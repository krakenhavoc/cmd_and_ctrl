package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2b3f2e85-5ea4-44d6-86e8-e8ce6d4f98c0",
		Name:            "Drowner of Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters("Drowner of Hope — create two 1/1 colorless Eldrazi Scions",
				Do(CreateToken{Template: b28EldraziScionToken(), N: 2})),
				game.Purpose{Tokens: 2, CreatureTokens: 2}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice an Eldrazi Scion: Tap target creature.",
			Cost:    SacrificeN(1, "an Eldrazi Scion", HasSubtype("Eldrazi"), HasSubtype("Scion")),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
