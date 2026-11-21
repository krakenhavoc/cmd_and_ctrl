package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "7da19fc5-cb85-4ec7-beea-95210378bbfe",
		Name:            "Balmor, Battlemage Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Balmor, Battlemage Captain — creatures you control get +1/+0 and gain trample until end of turn",
				effects.Do(
					effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Balmor, Battlemage Captain — +1/+0"},
					effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"trample"},
						Label: "Balmor, Battlemage Captain — trample"},
				)),
		},
	})
}
