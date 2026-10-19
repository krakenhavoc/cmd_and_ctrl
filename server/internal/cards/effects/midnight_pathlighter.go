package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Pathlighter — Creature — Human Wizard {2}{W}{U}, 2/3:
//
//	"Creatures you control can't be blocked except by legendary
//	 creatures.
//	 Whenever one or more creatures you control deal combat damage to
//	 a player, venture into the dungeon."
//
// The evasion is a printed block rule over the creatures its controller
// controls, read live; the trigger is once per damaged player per
// combat damage step (CR 603.2c, the Professional Face-Breaker shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e49228d6-6488-4cbc-aa53-303afd3a549b",
		Name:         "Midnight Pathlighter",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnMatching(And(Creature(), YouControl())), Legendary(), "legendary creatures"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil, ventureLabel("Midnight Pathlighter"), Venture),
		},
	})
}
