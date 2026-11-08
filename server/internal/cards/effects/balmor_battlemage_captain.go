package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balmor, Battlemage Captain — Legendary Creature — Bird Wizard {U}{R},
// 1/3 (Foundations, #4465):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, creatures you control
//	 get +1/+0 and gain trample until end of turn."
//
// A cast trigger (Archmage of Runes' WheneverYouCast over instants and
// sorceries), so a copied spell triggers nothing. The creatures are the
// ones you control as it resolves (CR 611.2c): one that arrives later in
// the turn gets neither half. The pump (layer 7c) and the trample (layer
// 6) are Decimator of the Provinces' two primitives, and two spells give
// +2/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7da19fc5-cb85-4ec7-beea-95210378bbfe",
		Name:            "Balmor, Battlemage Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Balmor, Battlemage Captain — creatures you control get +1/+0 and gain trample until end of turn",
				Do(
					BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Balmor, Battlemage Captain — +1/+0"},
					GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"trample"},
						Label: "Balmor, Battlemage Captain — trample"},
				)),
		},
	})
}
