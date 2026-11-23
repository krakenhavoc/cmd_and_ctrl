package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devoted Paladin — Creature — Orc Knight {4}{W}, 4/4 (Adventures in the
// Forgotten Realms):
//
//	"Beacon of Hope — When this creature enters, creatures you control
//	 get +1/+1 and gain vigilance until end of turn."
//
// "Beacon of Hope" is an ability word with no rules meaning (CR 207.2c).
// The creatures it affects are fixed as the trigger resolves (CR 611.2c),
// the Paladin among them, so a creature that enters later gets nothing.
//
// No simplification.
func init() {
	you := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "00bcbeb5-196a-4a54-ac8d-dec708e68d01",
		Name:         "Devoted Paladin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Devoted Paladin — creatures you control get +1/+1 and gain vigilance until end of turn",
				effects.Do(
					effects.BoostUntilEOT{Match: you, Power: 1, Toughness: 1, Label: "Devoted Paladin — +1/+1"},
					effects.GrantKeywordUntilEOT{Match: you, Keywords: []string{"vigilance"}, Label: "Devoted Paladin — vigilance"},
				)),
		},
	})
}
