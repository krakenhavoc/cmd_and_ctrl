package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Island of Wak-Wak — Land (Arabian Nights, #5641):
//
//	"{T}: Target creature with flying has base power 0 until end of
//	 turn."
//
// Singing Tree's layer 7b base power 0 on any creature with flying,
// whoever controls it. The land has no mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d427e61d-5b30-4d2a-bad2-2e7f016036ca",
		Name:         "Island of Wak-Wak",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature with flying has base power 0 until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.BasePowerZeroUntilEOT("Island of Wak-Wak — base power 0 until end of turn"),
		}},
	})
}
