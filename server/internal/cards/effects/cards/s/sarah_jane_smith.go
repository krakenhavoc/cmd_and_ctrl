package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarah Jane Smith — Legendary Creature — Human Detective {1}{W}, 2/1 (Doctor
// Who, #5948):
//
//	"Whenever you cast a historic spell, investigate. This ability triggers
//	 only once each turn. (Artifacts, legendaries, and Sagas are historic.)
//	 Doctor's companion"
//
// A cast trigger over a historic spell (CR 700.6, read off the spell as it
// was cast), once each turn (TriggersOnlyOnceEachTurn, used up when it
// triggers). Doctor's companion is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da268892-b190-4eee-9a2b-b0b9bb763033",
		Name:         "Sarah Jane Smith",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.WheneverYouCast(effects.B14Historic(),
				"Sarah Jane Smith — investigate", effects.Do(effects.Investigate{N: 1}))),
		},
	})
}
