package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gods' Eye, Gate to the Reikai — Legendary Land (Betrayers of Kamigawa,
// #5602):
//
//	"{T}: Add {C}.
//	 When Gods' Eye is put into a graveyard from the battlefield, create a
//	 1/1 colorless Spirit creature token."
//
// Dunes of the Dead's shape: "put into a graveyard from the battlefield"
// is the land's own dies-style leave trigger (WhenThisDies), whatever
// put it there — a sacrifice, a destruction or the legend rule (CR
// 704.5j). The token is the shared 1/1 colorless Spirit row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a66008c9-1ede-4dcf-8d35-6c0ed2390996",
		Name:         "Gods' Eye, Gate to the Reikai",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Gods' Eye, Gate to the Reikai — create a 1/1 colorless Spirit creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1})),
		},
	})
}
