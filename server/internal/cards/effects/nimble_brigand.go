package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nimble Brigand — Creature — Human Rogue {2}{U}, 1/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"This creature can't be blocked if you've committed a crime this
//	 turn. (Targeting opponents, anything they control, and/or cards in
//	 their graveyards is a crime.)
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// The evasion is a block rule on the creature itself (CantBeBlockedWhile,
// CR 509.1b), checked as blockers are declared against this turn's event
// log (YouCommittedACrimeThisTurn): a crime committed before blockers, even
// during combat, makes it unblockable. The draw is the ordinary
// combat-damage trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b153452-b4f2-4310-99e2-c729f28be6b9",
		Name:         "Nimble Brigand",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				return YouCommittedACrimeThisTurn(g, you)
			}),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Nimble Brigand — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
