package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "da268892-b190-4eee-9a2b-b0b9bb763033",
		Name:         "Sarah Jane Smith",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(WheneverYouCast(b14Historic(),
				"Sarah Jane Smith — investigate", Do(Investigate{N: 1}))),
		},
	})
}
